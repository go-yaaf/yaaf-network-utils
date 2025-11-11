package utils

import (
	"fmt"
	"github.com/ip2location/ip2location-io-go/ip2locationio"
	_ "github.com/ip2location/ip2location-io-go/ip2locationio"
)

type IPUtilsStruct struct {
	apiKey string
}

// IPUtils is a factory method that acts as a static member
func IPUtils() *IPUtilsStruct {
	return &IPUtilsStruct{
		apiKey: "A804D17F1EE16FBE269FE00610B95C97",
	}
}

func (t *IPUtilsStruct) GeoLookupWKT(ip string) (string, error) {
	config, err := ip2locationio.OpenConfiguration(t.apiKey)
	if err != nil {
		return "", err
	}
	ipl, err := ip2locationio.OpenIPGeolocation(config)
	if err != nil {
		return "", err
	}

	res, err := ipl.LookUp(ip, "") // language parameter only available with Plus and Security plans
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("POINT(%f %f)", res.Longitude, res.Latitude), nil
}

func (t *IPUtilsStruct) AddressLookup(ip string) (string, error) {
	config, err := ip2locationio.OpenConfiguration(t.apiKey)
	if err != nil {
		return "", err
	}
	ipl, err := ip2locationio.OpenIPGeolocation(config)
	if err != nil {
		return "", err
	}

	res, err := ipl.LookUp(ip, "")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("[%s], %s, %s, %s %s)", res.AS, res.CityName, res.RegionName, res.CountryName, res.ZipCode), nil
}
