# Paperless-launcher

A docker container used to mount/unmount an encypted [veracrypt](https://www.veracrypt.fr/en/Home.html) volume and start [paperless](https://github.com/paperless-ngx/paperless-ngx) on demand.

## Usage

Edit the volume line in the `docker-compose.yml` file to map your folder with your veracrypt volumes, and the traefik-related settings.
Then run `docker-compose up -d` to start the container.

|CLI flag|Env var|Default|Description|
|-----|---|-------|-----------|
|serve|PLL_SERVE|0.0.0.0:3000|the bind address of the server|
|user-header|PLL_USER_HEADER|Remote-User|the header used to authenticate the user|
|email-header|PLL_EMAIL_HEADER|Remote-Email|the header that contains the user's email address|
|registration|PLL_REGISTRATION|false|Enable auto creation of volumes for new users (CURRENTLY NOT AVAILABLE)|
|size|PLL_SIZE|2G|The default size of a veracrypt volume for the auto creation|
|mount-path|PLL_MOUNT_PATH|/app/mount/%user%|the path where the veracrypt volume will be mounted, `%user%` will be repaced by the username
|volume-path|PLL_VOLUME_PATH|/data/%user%.hc|the path to the veracrypt volume for a specific user, this can be overwritten in the `mappings.json` file with `"youruser": "yourpath"`|
|mapping|PLL_MAPPING|mapping.json|Path to the user to path mapping file|
|timeout|PLL_TIMEOUT|1|Inactivity timeout before unmounting the volume (in minutes)|
|startPort|PLL_START_PORT|10000|First port to be used to bind the paperless container|
|redis-image|PLL_REDIS_IMAGE|redis:latest|the redis docker image to use|
|paperless-image|PLL_PAPERLESS_IMAGE|ghcr.io/paperless-ngx/paperless-ngx:latest|the paperless docker image to use|

## How it works

This is a simple container running dind (docker-in-docker) with veracrypt.

Each user have his own veracrypt volume that stores his documents and the paperless database (sqlite).

When the user wants to login, the request is handled by the forward-auth of the reverse-proxy (here, [authelia](https://github.com/authelia/authelia)) that adds a `Remote-User` header which is used to authenticate the user.

Then this application will ask the user for the password of the veracrypt volume to decrypt it, if it is not mounted and will mount the volume, start a paperless container and redirect the user to his paperless instance.

When the user have finished, he can click on logout or wait for the inactivity timeout, after this, the application will stop the paperless container and unmount the volume.

Note that this is not a 100% secure since the files remains unencrypted when you are accessing paperless, but one you are finished, the volume is unmounted, and the files cannot be accessed without the volume password. It is of course strongly recommanded to use https to connect to this application, as the volume password could be intercepted if only using http.


