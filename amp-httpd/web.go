package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewWeb(amp *Amp) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(setZone)
	e.GET("/zones/:id/power", handleGetZonePower(amp))
	e.PUT("/zones/:id/power", handleSetZonePower(amp))
	e.GET("/zones/:id/volume", handleGetVolume(amp))
	e.PUT("/zones/:id/volume", handleSetZoneVolume(amp))
	e.GET("/zones/:id/source", handleGetSource(amp))
	e.PUT("/zones/:id/source", handleSetZoneSource(amp))

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

func ignoreAmpOff(err error) error {
	if _, ok := err.(*AmpOffError); ok {
		return nil
	}
	return err
}