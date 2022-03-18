package main

import (
	"log"
	"os/exec"
	"strconv"
	"time"
)

func pullImages() error {
	log.Println("pulling " + config.RedisImage)
	cmd := exec.Command("docker", "pull", config.RedisImage)
	data, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %s", string(data))
		return err
	}

	log.Println("pulling " + config.PaperlessImage)
	cmd = exec.Command("docker", "pull", config.PaperlessImage)
	data, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %s", string(data))
		return err
	}
	return nil
}

func killPaperless(user string) error {
	data, ok := users[user]
	if ok {
		log.Printf("stopping paperless for %s", user)

		cmd := exec.Command("docker", "stop", data.DockerPaperless)
		err := cmd.Run()
		if err != nil {
			return err
		}

		cmd = exec.Command("docker", "stop", data.DockerRedis)
		err = cmd.Run()
		if err != nil {
			return err
		}

		cmd = exec.Command("docker", "network", "rm", data.DockerNetwork)
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func contains(slice []int, elem int) bool {
	for _, x := range slice {
		if x == elem {
			return true
		}
	}
	return false
}

func getUnusedPort(user string) int {
	ports := []int{}

	for _, v := range users {
		ports = append(ports, v.Port)
	}

	p := config.StartPort
	for true {
		if !contains(ports, p) {
			return p
		}
		p++
	}
	return -1
}

func spawnPaperless(volumePath string, mountPath string, user string, password string, email string) (User, error) {
	killPaperless(user)

	redis := "pll_" + user + "_redis"
	network := "pll_" + user + "_net"
	paperless := "pll_" + user + "_paperless"
	port := getUnusedPort(user)

	log.Printf("starting paperless for %s", user)

	cmd := exec.Command("docker", "network", "create", network)
	data, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %s", string(data))
	}

	cmd = exec.Command("docker", "run", "--rm", "-d", "--network="+network, "--log-driver=none", "--name="+redis, config.RedisImage)
	data, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %s", string(data))
		return User{}, err
	}

	env := map[string]string{
		"PAPERLESS_REDIS":               "redis://" + redis + ":6379",
		"PAPERLESS_ADMIN_USER":          user,
		"PAPERLESS_ADMIN_PASSWORD":      password,
		"PAPERLESS_ADMIN_MAIL":          email,
		"PAPERLESS_AUTO_LOGIN_USERNAME": user,
		"PAPERLESS_FORCE_SCRIPT_NAME":   "/" + user,
		"PAPERLESS_COOKIE_PREFIX":       "pll_" + user + "_",
	}
	params := []string{"run", "--rm", "-d", "-p", "127.0.0.1:" + strconv.Itoa(port) + ":8000", "--network=" + network, "--log-driver=none", "--name=" + paperless, "-v", mountPath + "/data:/usr/src/paperless/data", "-v", mountPath + "/media:/usr/src/paperless/media"}
	for k, v := range env {
		params = append(params, "-e")
		params = append(params, k+"="+v)
	}
	for k, v := range config.PaperlessEnv {
		if _, ok := env[k]; !ok {
			params = append(params, "-e")
			params = append(params, k+"="+v)
		}
	}

	params = append(params, config.PaperlessImage)
	cmd = exec.Command("docker", params...)
	data, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %s", string(data))
		return User{}, err
	}

	timeout := time.AfterFunc(config.SessionTimeout, func() {
		log.Printf("timeout for %s", user)
		logoutUser(user)
	})

	return User{Port: port, DockerNetwork: network, DockerPaperless: paperless, DockerRedis: redis, Timeout: timeout, VolumePath: volumePath, MountPath: mountPath}, nil
}
