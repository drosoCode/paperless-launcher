package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func getMountPath(user string) string {
	return strings.Replace(config.MountPath, "%user%", user, -1)
}

func mount(user string, password string) error {
	mapping := map[string]string{}

	// check that mount path exists
	mnt := getMountPath(user)
	if _, err := os.Stat(mnt); os.IsNotExist(err) {
		if err := os.MkdirAll(mnt, 0755); err != nil {
			return err
		}
	}

	// read mappings file
	if _, err := os.Stat(config.MappingFile); err == nil {
		jdata, err := os.ReadFile(config.MappingFile)
		if err == nil {
			err = json.Unmarshal(jdata, &mapping)
		}
	}

	// if user in mappings
	if path, ok := mapping[user]; ok {
		if _, err := os.Stat(path); err != nil {
			// if volume does not exists, create it
			if !config.EnableRegistration {
				return errors.New("registration is disabled")
			}
			err := createVeracryptVolume(path, password)
			if err != nil {
				return err
			}
		}
		return mountVeracryptVolume(path, mnt, password)
	}

	// else, default to config.DataPath
	dataPath := strings.Replace(config.VolumePath, "%user%", user, -1)

	// if volume does not exists, create it
	if _, err := os.Stat(dataPath); err != nil {
		if !config.EnableRegistration {
			return errors.New("registration is disabled")
		}
		err := createVeracryptVolume(dataPath, password)
		if err != nil {
			return err
		}
	}

	return mountVeracryptVolume(dataPath, mnt, password)
}

func unmount(user string) error {
	data, ok := users[user]
	if ok {
		log.Println("unmounting volume")
		cmd := exec.Command("/bin/sh", "-c", "cd "+path.Dir(data.MountPath)+" && veracrypt -d "+path.Base(data.MountPath))
		return cmd.Run()
	}
	return nil
}

func mountVeracryptVolume(volumePath string, mountPath string, password string) error {
	log.Println("mounting volume")
	cmd := exec.Command("veracrypt", volumePath, mountPath, "--pim=0", "-k", "", "--protect-hidden=no", "--non-interactive", "-m=nokernelcrypto", "--stdin")
	cmd.Stdin = strings.NewReader(password)
	data, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %s", string(data))
		return err
	}
	return nil
}

func createVeracryptVolume(volumePath string, password string) error {
	log.Println("creating volume")
	cmd := exec.Command("veracrypt", "-m=nokernelcrypto", "--create", "--volume-type=normal", "--size=1G", "--encryption=AES", "--hash=SHA-512", "--filesystem=Ext4", "--pim=0", "-k", "", "--random-source=/dev/urandom", "--non-interactive", "--stdin", volumePath)
	cmd.Stdin = strings.NewReader(password)
	data, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %s", string(data))
		return err
	}
	cmd = exec.Command("/bin/sh", "-c", "chown 0:100 "+volumePath)
	data, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %s", string(data))
		return err
	}
	return nil
}
