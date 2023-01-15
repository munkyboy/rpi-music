package main

import (
	"bufio"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"sync"

	"github.com/tarm/serial"
)

type Amp struct {
	*serial.Port
	mutex sync.Mutex
}

type AmpOffError struct{}

func (m *AmpOffError) Error() string {
	return "amp is off"
}

// real max is 38 but we don't want to blow speakers
const MaxVolume uint = 20

func NewAmp(d *serial.Port) *Amp {
	return &Amp{d, sync.Mutex{}}
}

func (amp *Amp) GetVolume(zone uint) (uint, error) {
	if v, err := amp.sendSingleZoneInquiry(fmt.Sprintf("?%dVO\r", zone)); err != nil {
		return 0, err
	} else {
		return uint(v), nil
	}
}

func (amp *Amp) SetVolume(zone uint, volume uint) error {
	if volume > MaxVolume {
		return fmt.Errorf("volume must be 0 - %d", MaxVolume)
	}
	if err := amp.sendCommand(fmt.Sprintf("<%dVO%02d\r", zone, volume)); err != nil {
		return err
	}
	return nil
}

func (amp *Amp) GetSource(zone uint) (uint, error) {
	if v, err := amp.sendSingleZoneInquiry(fmt.Sprintf("?%dCH\r", zone)); err != nil {
		return 0, err
	} else {
		return uint(v), nil
	}
}

func (amp *Amp) SetSource(zone uint, source uint) error {
	if source < 1 || source > 6 {
		return fmt.Errorf("source must be between 1 and 6")
	}
	if err := amp.sendCommand(fmt.Sprintf("<%dCH%02d\r", zone, source)); err != nil {
		return err
	}
	return nil
}

func (amp *Amp) GetPower(zone uint) (bool, error) {
	if v, err := amp.sendSingleZoneInquiry(fmt.Sprintf("?%dPR\r", zone)); err != nil {
		return false, err
	} else if v == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (amp *Amp) SetPower(zone uint, power bool) error {
	value := "00"
	if power {
		value = "01"
	}
	if err := amp.sendCommand(fmt.Sprintf("<%dPR%s\r", zone, value)); err != nil {
		return err
	}
	return nil
}

func (amp *Amp) sendCommand(cmd string) error {
	amp.mutex.Lock()
	defer amp.mutex.Unlock()
	amp.Flush()
	if _, err := amp.Write([]byte(cmd)); err != nil {
		return err
	} else if _, err := amp.readReply(len(cmd)); err != nil {
		return err
	}
	return nil
}

func (amp *Amp) sendSingleZoneInquiry(cmd string) (int, error) {
	amp.mutex.Lock()
	defer amp.mutex.Unlock()
	if _, err := amp.Write([]byte(cmd)); err != nil {
		return 0, err
	} else if r, err := amp.readReply(len(cmd)); err != nil {
		return 0, err
	} else if v, err := parseSingleZoneReply(r); err != nil {
		return 0, err
	} else {
		return v, nil
	}
}

var errorPattern *regexp.Regexp = regexp.MustCompile(`Command Error`)

// The amp will echo back the command followed by a prompt so we need to know
// the length of the command in order to extract the reply
func (amp *Amp) readReply(cmdLen int) (reply string, err error) {
	scanner := bufio.NewScanner(amp)
	for scanner.Scan() {
		reply = reply + scanner.Text() + "\n"
	}
	// fmt.Printf("got reply: %q\n", reply)
	// when amp is off, our serial library is configured to timeout and will
	// return an empty string. return an AppOffError in this case.
	if len(reply) == 0 {
		err = new(AmpOffError)
		return
	}

	if err = scanner.Err(); err != nil {
		return
	} else if errorPattern.MatchString(reply) {
		err = fmt.Errorf("Command Error")
		return
	}

	if len(reply) > cmdLen {
		// The +1 accounts for the prompt char (`#`)
		reply = reply[cmdLen + 1:]
	}
	return
}

var singleZoneReplyFormat *regexp.Regexp = regexp.MustCompile(`^>\d{2}\w{2}(\d{2})\r`)

func parseSingleZoneReply(b string) (int, error) {
	if !singleZoneReplyFormat.MatchString(b) {
		return 0, fmt.Errorf("unknown reply format: %q", b)
	}
	matches := singleZoneReplyFormat.FindStringSubmatch(b)
	v, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}
	return v, nil
}

func VolumeFromPercentage(v uint) uint {
	if v > 100 {
		return MaxVolume
	}
	return uint(math.Round(float64(v) / 100 * float64(MaxVolume)))
}

func VolumeToPercentage(v uint) uint {
	if v > MaxVolume {
		return 100
	}
	return uint(math.Round(float64(v) / float64(MaxVolume) * 100))
}