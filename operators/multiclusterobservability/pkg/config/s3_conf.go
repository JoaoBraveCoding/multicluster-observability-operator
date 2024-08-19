// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project
// Licensed under the Apache License 2.0

package config

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v2"
)

func validateS3(conf Config) error {

	if conf.Bucket == "" {
		return errors.New("no s3 bucket in config file")
	}

	if conf.Endpoint == "" {
		return errors.New("no s3 endpoint in config file")
	}

	return nil
}

// IsValidS3Conf is used to validate s3 configuration.
func IsValidS3Conf(data []byte) error {
	var objectConfg ObjectStorgeConf
	err := yaml.Unmarshal(data, &objectConfg)
	if err != nil {
		return err
	}

	if strings.ToLower(objectConfg.Type) != "s3" {
		return errors.New("invalid type config, only s3 type is supported")
	}

	err = validateS3(objectConfg.Config)
	if err != nil {
		return err
	}

	return nil
}
