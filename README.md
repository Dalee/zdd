##
[![Build Status](https://travis-ci.org/arkady-emelyanov/zdd.svg?branch=master)](https://travis-ci.org/arkady-emelyanov/zdd)

## Why?
Use docker for new projects is simple. For old, complex projects setup cloud environment
and configure all required service discovery things can take forever. Otherwise,
for a small projects deployed on one host, use of full power of cloud is not required.

To allow use power of Docker in such environments, here it is: `zdd`:
 * Create and start new container from image
 * Start and bootstrap freshly created container
 * Check container is alive for some period of time
 * Update upstream proxy config and reload upstream

In other words, it can deploy new container with zero downtime! Vagrant-powered
playground is included, check it out:

```bash
$ vagrant up
```

Point your browser to localhost:8080 and observe `502 Bad Gateway` - it's ok,
because no version is deployed, yet. Just run:

```bash
$ vagrant ssh
$ glide install
$ go run ./main.go deploy -c ./example.yml -v 1.11
```

Wait until deploy is done and refresh browser. Magic.

## Basic usage

Deploy new version
```bash
$ zdd deploy -c example.yml -v 1.11
```

Rollback to previous version (should be at least two releases):
```bash
$ zdd rollback -c example.yml
```

Tool will store deploy log under `${HOME}/.zdd/<name>.deploy_log`, so please make
sure user have home and it's allowed to write.

## Sample configuration file

```yml
# name of image to use, mandatory
image: nginx

# prefix for container names, final name will be:
# <name>.v<tag>.<millisecond>
name: nginx-staging

# set of environment variables to set to container
env:
  - HELLO=WORLD

# define ports exposed by container
# each port available in upstream template as:
# 80/tcp => %TCP_80%
# 514/udp => %UDP_514%
port:
  - 80/tcp

# list of commands to run inside of container
# right after container is created, but
# before upstream config is switched
# useful for migrations / post-deploy things
bootstrap:
  - touch /root/.alive

# list of mounts
# mount directory to container local_path:container_path
mount:
  - /home/vagrant/html:/usr/share/nginx/html

# locally installed upstream which is set up
# in front of docker
upstream:
  - local_nginx:
    command: "sudo service nginx reload"
    resource: /home/vagrant/conf.d/nginx_staging_upstream.conf
    template: |
      upstream nginx_staging {
        server 127.0.0.1:%TCP_80%;
      }
```

Configuration file placeholders:
* Upstream template only: `%TCP_<PORT_NUMBER>%` - put auto-generated value of
port number in "port" array, default is TCP, can be UDP
* Any configuration value: `${<ANY>}` - put value of environment variable `ANY`

`zdd` will try to connect to Docker host defined in `${DOCKER_HOST}` or default
platform docker socket and use default API version or version defined
in `${DOCKER_API_VERSION}` environment variable.

Work currently in progress, so.. *use at your own risk*
