package parser

import (
	"fmt"
	"net"
	"strconv"

	"github.com/crowdsecurity/crowdsec/pkg/types"
	log "github.com/sirupsen/logrus"

	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
)

func IpToRange(field string, p *types.Event, ctx interface{}) (map[string]string, error) {
	var dummy interface{}
	ret := make(map[string]string)

	if field == "" {
		return nil, nil
	}
	ip := net.ParseIP(field)
	if ip == nil {
		log.Infof("Can't parse ip %s, no range enrich", field)
		return nil, nil
	}
	net, ok, err := ctx.(*maxminddb.Reader).LookupNetwork(ip, &dummy)
	if err != nil {
		log.Errorf("Failed to fetch network for %s : %v", ip.String(), err)
		return nil, nil
	}
	if !ok {
		log.Debugf("Unable to find range of %s", ip.String())
		return nil, nil
	}
	ret["SourceRange"] = net.String()
	return ret, nil
}

func GeoIpASN(field string, p *types.Event, ctx interface{}) (map[string]string, error) {
	ret := make(map[string]string)
	if field == "" {
		return nil, nil
	}

	ip := net.ParseIP(field)
	if ip == nil {
		log.Infof("Can't parse ip %s, no ASN enrich", ip)
		return nil, nil
	}
	record, err := ctx.(*geoip2.Reader).ASN(ip)
	if err != nil {
		log.Errorf("Unable to enrich ip '%s'", field)
		return nil, nil
	}
	ret["ASNNumber"] = fmt.Sprintf("%d", record.AutonomousSystemNumber)
	ret["ASNOrg"] = record.AutonomousSystemOrganization

	log.Tracef("geoip ASN %s -> %s, %s", field, ret["ASNNumber"], ret["ASNOrg"])

	return ret, nil
}

func GeoIpCity(field string, p *types.Event, ctx interface{}) (map[string]string, error) {
	ret := make(map[string]string)
	if field == "" {
		return nil, nil
	}
	ip := net.ParseIP(field)
	if ip == nil {
		log.Infof("Can't parse ip %s, no City enrich", ip)
		return nil, nil
	}
	record, err := ctx.(*geoip2.Reader).City(ip)
	if err != nil {
		log.Debugf("Unable to enrich ip '%s'", ip)
		return nil, nil
	}
	ret["IsoCode"] = record.Country.IsoCode
	ret["IsInEU"] = strconv.FormatBool(record.Country.IsInEuropeanUnion)
	ret["Latitude"] = fmt.Sprintf("%f", record.Location.Latitude)
	ret["Longitude"] = fmt.Sprintf("%f", record.Location.Longitude)

	log.Tracef("geoip City %s -> %s, %s", field, ret["IsoCode"], ret["IsInEU"])

	return ret, nil
}

func GeoIPCityInit(cfg map[string]string) (interface{}, error) {
	dbCityReader, err := geoip2.Open(cfg["datadir"] + "/GeoLite2-City.mmdb")
	if err != nil {
		log.Debugf("couldn't open geoip : %v", err)
		return nil, err
	}

	return dbCityReader, nil
}

func GeoIPASNInit(cfg map[string]string) (interface{}, error) {
	dbASReader, err := geoip2.Open(cfg["datadir"] + "/GeoLite2-ASN.mmdb")
	if err != nil {
		log.Debugf("couldn't open geoip : %v", err)
		return nil, err
	}

	return dbASReader, nil
}

func IpToRangeInit(cfg map[string]string) (interface{}, error) {
	ipToRangeReader, err := maxminddb.Open(cfg["datadir"] + "/GeoLite2-ASN.mmdb")
	if err != nil {
		log.Debugf("couldn't open geoip : %v", err)
		return nil, err
	}

	return ipToRangeReader, nil
}
