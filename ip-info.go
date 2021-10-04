package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"
)

type IPInfo struct {
	IP            string  `json:"ip"`
	Type          string  `json:"type"`
	ContinentCode string  `json:"continent_code,omitempty"`
	ContinentName string  `json:"continent_name,omitempty"`
	CountryCode   string  `json:"country_code,omitempty"`
	CountryName   string  `json:"country_name,omitempty"`
	RegionCode    string  `json:"region_code,omitempty"`
	RegionName    string  `json:"region_name,omitempty"`
	City          string  `json:"city,omitempty"`
	Zip           string  `json:"zip,omitempty"`
	Latitude      float64 `json:"latitude,omitempty"`
	Longitude     float64 `json:"longitude,omitempty"`
	Location      struct {
		GeonameID int    `json:"geoname_id,omitempty"`
		Capital   string `json:"capital,omitempty"`
		Languages []struct {
			Code   string `json:"code,omitempty"`
			Name   string `json:"name,omitempty"`
			Native string `json:"native,omitempty"`
		} `json:"languages,omitempty"`
		CountryFlag             string `json:"country_flag,omitempty"`
		CountryFlagEmoji        string `json:"country_flag_emoji,omitempty"`
		CountryFlagEmojiUnicode string `json:"country_flag_emoji_unicode,omitempty"`
		CallingCode             string `json:"calling_code,omitempty"`
		IsEu                    bool   `json:"is_eu,omitempty"`
	} `json:"location,omitempty"`
}

func (ip *IPInfo) JSONString() string {
	empJSON, _ := json.MarshalIndent(ip, "", "  ")
	return string(empJSON)
}

func (ip *IPInfo) JSONBytes() []byte {
	jsonByte, _ := json.MarshalIndent(ip, "", "  ")
	return jsonByte
}

func (ip *IPInfo) MessageString() string {
	message := "<code>IP:</code> " + ip.IP
	message += "\n<code>Type:</code> " + ip.Type
	message += "\n<code>Continent:</code> " + ip.ContinentName
	message += "\n<code>Country:</code> " + ip.CountryName + " " + ip.Location.CountryFlagEmoji
	message += "\n<code>Region:</code> " + ip.RegionName
	message += "\n<code>City:</code> " + ip.City
	return message
}

func getIPInfo(ip net.IP) (*IPInfo, error) {
	request, err := http.NewRequest(http.MethodGet, os.Getenv("IPSTACK_URL")+ip.String(), nil)
	if err != nil {
		return nil, err
	}

	query := request.URL.Query()
	query.Add("access_key", os.Getenv("IPSTACK_ACCESS_KEY"))
	request.URL.RawQuery = query.Encode()

	client := http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	IPData := IPInfo{}
	err = json.Unmarshal(responseBytes, &IPData)
	if err != nil {
		return nil, err
	}

	return &IPData, nil
}
