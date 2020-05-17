package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/covista/commons/internal/config"
	"github.com/covista/commons/proto"
)

func checkConfig(cfg *config.Config) error {
	if cfg == nil {
		return errors.New("Configuration is nil")
	} else if len(cfg.Database.Host) == 0 {
		return errors.New("Database.Host is empty")
	} else if len(cfg.Database.Database) == 0 {
		return errors.New("Database.Database is empty")
	} else if len(cfg.Database.User) == 0 {
		return errors.New("Database.User is empty")
	} else if len(cfg.Database.Password) == 0 {
		return errors.New("Database.Password is empty")
	} else if len(cfg.Database.Port) == 0 {
		return errors.New("Database.Port is empty")
	}
	return nil
}

func checkTokenRequest(req *proto.TokenRequest) error {
	if req == nil {
		return errors.New("Empty TokenRequest")
	} else if req.ApiKey == nil {
		return errors.New("Empty TokenRequest api_key")
	} else if len(req.ApiKey) != 16 {
		return errors.New("api_key is not correct length")
	} else if !parsesAsRFC3339(req.PermittedRangeStart) {
		return errors.New("permitted_range_start is not an RFC3339-formatted timestamp")
	} else if !parsesAsRFC3339(req.PermittedRangeEnd) {
		return errors.New("permitted_range_end is not an RFC3339-formatted timestamp")
	} else {
		return nil
	}
}

func checkReport(rep *proto.Report) error {
	if rep == nil {
		return errors.New("Empty Report")
	} else if rep.AuthorizationKey == nil {
		return errors.New("Empty AuthorizationKey")
	} else if len(rep.AuthorizationKey) != 16 {
		return errors.New("authorization_key is not correct length")
	} else if len(rep.Reports) == 0 {
		return errors.New("report does not contain any reports")
	}
	for idx, report := range rep.Reports {
		if err := checkTimestampedTEK(report); err != nil {
			return fmt.Errorf("Report %d is invalid: %w", idx, err)
		}
	}
	return nil
}

func checkTimestampedTEK(tek *proto.TimestampedTEK) error {
	if len(tek.TEK) != 16 {
		return errors.New("TEK was invalid length")
	} else if tek.ENIN == 0 {
		return errors.New("ENIN was invalid")
	}
	return nil
}

func checkGetKeyRequest(req *proto.GetKeyRequest) error {
	if req == nil {
		return errors.New("Empty query")
	} else if len(req.HAK) == 0 && req.ENIN == 0 && req.Hrange == nil {
		return errors.New("GetKeyRequest does not define any filters")
	} else if req.Hrange != nil && (len(req.Hrange.StartDate) == 0 && req.Hrange.Days == 0) {
		return errors.New("GetKeyRequest.historical_range is empty")
	}
	return nil
}

func parsesAsRFC3339(ts string) bool {
	_, err := time.Parse(time.RFC3339, ts)
	return err == nil
}
