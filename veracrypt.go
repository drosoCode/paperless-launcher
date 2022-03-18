package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
)

func getMountPath(user string) string {
	return strings.Replace(config.MountPath, "%user%", user, -1)
}

func mount(user string, password string) (string, string, error) {
	mapping := map[string]string{}

	// check that mount path exists
	mnt := getMountPath(user)
	if _, err := os.Stat(mnt); os.IsNotExist(err) {
		if err := os.MkdirAll(mnt, 0755); err != nil {
			return "", "", err
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
				return "", "", errors.New("registration is disabled")
			}
			err := createVeracryptVolume(path, password)
			if err != nil {
				return "", "", err
			}
		}
		return path, mnt, mountVeracryptVolume(path, mnt, password)
	}

	// else, default to config.DataPath
	dataPath := strings.Replace(config.VolumePath, "%user%", user, -1)

	// if volume does not exists, create it
	if _, err := os.Stat(dataPath); err != nil {
		if !config.EnableRegistration {
			return "", "", errors.New("registration is disabled")
		}
		err := createVeracryptVolume(dataPath, password)
		if err != nil {
			return "", "", err
		}
	}

	return dataPath, mnt, mountVeracryptVolume(dataPath, mnt, password)
}

func unmount(volumePath string) error {
	log.Println("unmounting volume")
	cmd := exec.Command("veracrypt", "-d", volumePath)
	return cmd.Run()
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
	// create volume
	cmd := exec.Command("veracrypt", "--create", "--volume-type=normal", "--size="+config.DefaultVolumeSize, "--encryption=AES", "--hash=SHA-512", "--filesystem=none", "--pim=0", "-k", "", "--random-source=/dev/urandom", "--non-interactive", "--stdin", volumePath)
	cmd.Stdin = strings.NewReader(password)
	data, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %s", string(data))
		return err
	}
	// mount volume with empty fs
	cmd = exec.Command("veracrypt", "-m=nokernelcrypto", "--non-interactive", "--filesystem=none", "--pim=0", "-k", "", "--protect-hidden=no", "--stdin", volumePath)
	cmd.Stdin = strings.NewReader(password)
	data, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %s", string(data))
		return err
	}
	// create fs
	cmd = exec.Command("mkfs.ext4", "/tmp/.veracrypt_aux_mnt1/volume")
	data, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %s", string(data))
	}
	// unmount
	return unmount(volumePath)
}
