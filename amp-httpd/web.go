package main

import (
	"fmt"
	"net/http"
	"os/exec"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewWeb(amp *Amp) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	zones := e.Group("/zones")
	zones.Use(setZone)
	zones.GET("/:id/power", handleGetZonePower(amp))
	zones.PUT("/:id/power", handleSetZonePower(amp))
	zones.GET("/:id/volume", handleGetVolume(amp))
	zones.PUT("/:id/volume", handleSetZoneVolume(amp))
	zones.GET("/:id/source", handleGetSource(amp))
	zones.PUT("/:id/source", handleSetZoneSource(amp))

	e.GET("/power", handleGetAmpPower(amp))
	e.PUT("/power", handleSetAmpPower(amp))
	return e
}

type zone struct {
	ID     uint `param:"id"`
	Power  bool `form:"power"`
	Volume uint `form:"volume"`
	Source uint `form:"source"`
}

func setZone(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		z := zone{}
		if err := c.Bind(&z); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		c.Set("zone", z)
		return next(c)
	}
}

func handleGetZonePower(amp *Amp) echo.HandlerFunc {
	return func(c echo.Context) error {
		z, _ := c.Get("zone").(zone)
		v, err := amp.GetPower(z.ID)
		if _, ok := err.(*AmpOffError); ok {
			return c.String(http.StatusOK, "false\n")
		} else if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, fmt.Sprintf("%t\n", v))
	}
}

func handleSetZonePower(amp *Amp) echo.HandlerFunc {
	return func(c echo.Context) error {
		z, _ := c.Get("zone").(zone)
		if err := ignoreAmpOff(amp.SetPower(z.ID, z.Power)); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "ok\n")
	}
}

func handleGetVolume(amp *Amp) echo.HandlerFunc {
	return func(c echo.Context) error {
		z, _ := c.Get("zone").(zone)
		v, err := amp.GetVolume(z.ID)
		if _, ok := err.(*AmpOffError); ok {
			return c.String(http.StatusOK, "0\n")
		} else if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, fmt.Sprintf("%d\n", v))
	}
}

func handleSetZoneVolume(amp *Amp) echo.HandlerFunc {
	return func(c echo.Context) error {
		z, _ := c.Get("zone").(zone)
		if err := ignoreAmpOff(amp.SetVolume(z.ID, z.Volume)); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "ok\n")
	}
}

func handleGetSource(amp *Amp) echo.HandlerFunc {
	return func(c echo.Context) error {
		z, _ := c.Get("zone").(zone)
		v, err := amp.GetSource(z.ID)
		if _, ok := err.(*AmpOffError); ok {
			return c.String(http.StatusOK, "1\n")
		} else if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, fmt.Sprintf("%d\n", v))
	}
}

func handleSetZoneSource(amp *Amp) echo.HandlerFunc {
	return func(c echo.Context) error {
		z, _ := c.Get("zone").(zone)
		if err := ignoreAmpOff(amp.SetSource(z.ID, z.Source)); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "ok\n")
	}
}

func handleGetAmpPower(amp *Amp) echo.HandlerFunc {
	return func(c echo.Context) error {
		// get power of any zone as a proxy for amp being on
		_, err := amp.GetPower(11)
		if _, ok := err.(*AmpOffError); ok {
			return c.String(http.StatusOK, "false\n")
		} else if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "true\n")
	}
}

func handleSetAmpPower(amp *Amp) echo.HandlerFunc {
	return func(c echo.Context) error {
		power := c.FormValue("power")
		cmd := exec.Command("/usr/local/bin/amp")
		if power == "0" {
			cmd.Args = append(cmd.Args, "off")
		} else if power == "1" {
			cmd.Args = append(cmd.Args, "on")
		} else {
			return echo.NewHTTPError(http.StatusBadRequest, "expecting power value of `0` or `1`")
		}
		if err := cmd.Run(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "ok\n")
	}
}

func ignoreAmpOff(err error) error {
	if _, ok := err.(*AmpOffError); ok {
		return nil
	}
	return err
}