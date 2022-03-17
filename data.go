package main

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/jamiealquiza/envy"
)

type User struct {
	Port            int
	DockerNetwork   string
	DockerPaperless string
	DockerRedis     string
	Timeout         *time.Timer
	MountPath       string
}

type Config struct {
	Serve              string
	RemoteUserHeader   string
	RemoteEmailHeader  string
	EnableRegistration bool
	DefaultVolumeSize  string
	VolumePath         string
	MountPath          string
	SessionTimeout     time.Duration
	PaperlessEnv       map[string]string
	MappingFile        string
	StartPort          int
	RedisImage         string
	PaperlessImage     string
}

func parseConfig() Config {
	serve := flag.String("serve", "0.0.0.0:3000", "bind address")
	remoteUser := flag.String("user-header", "Remote-User", "name of the user header")
	remoteEmail := flag.String("email-header", "Remote-Email", "name of the email header")
	//registration := flag.Bool("registration", false, "enable registration")
	size := flag.String("size", "2G", "size of the new volumes")
	mountPath := flag.String("mount-path", "./mount/%user%", "path to mount the veracrypt volumes, use %user% to replace by the username")
	volumePath := flag.String("volume-path", "./volumes/%user%.vc", "default path to create new volumes, use %user% to replace by the username")
	timeout := flag.Int("timeout", 1, "no activity timeout before relocking the volume, default: 10 min, 0 to disable")
	paperlessEnv := flag.String("paperless-env", "", "path to a json file with env params for paperless")
	mappingFile := flag.String("mapping", "mapping.json", "path to the user mapping file")
	startPort := flag.Int("start-port", 10000, "first port to be used to bind the container")
	redis := flag.String("redis-image", "redis:latest", "redis docker image")
	paperless := flag.String("paperless-image", "ghcr.io/paperless-ngx/paperless-ngx:latest", "paperless docker image")

	envy.Parse("PLL")
	flag.Parse()

	conf := Config{
		Serve:              *serve,
		RemoteUserHeader:   *remoteUser,
		RemoteEmailHeader:  *remoteEmail,
		EnableRegistration: false,
		DefaultVolumeSize:  *size,
		MountPath:          *mountPath,
		VolumePath:         *volumePath,
		SessionTimeout:     time.Duration(*timeout) * time.Minute,
		PaperlessEnv:       map[string]string{},
		MappingFile:        *mappingFile,
		StartPort:          *startPort,
		PaperlessImage:     *paperless,
		RedisImage:         *redis,
	}

	if _, err := os.Stat(*mappingFile); err != nil {
		os.WriteFile(conf.MappingFile, []byte("{}"), 0644)
	}

	if *paperlessEnv != "" {
		if _, err := os.Stat(*paperlessEnv); err == nil {
			jdata, err := os.ReadFile(*paperlessEnv)
			if err == nil {
				err = json.Unmarshal(jdata, &conf.PaperlessEnv)
			}
		}
	}

	return conf
}
