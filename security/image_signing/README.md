# Content trust in Docker

When transferring data among networked systems, trust is a central concern. In particular, when communicating over an untrusted medium such as the internet, it is critical to ensure the integrity and the publisher of all the data a system operates on. You use the Docker Engine to push and pull images (data) to a public or private registry. Content trust gives you the ability to verify both the integrity and the publisher of all the data received from a registry over any channel.

## Docker Content Trust (DCT)

Docker Content Trust (DCT) provides the ability to use digital signatures for data sent to and received from remote Docker registries. These signatures allow client-side or runtime verification of the integrity and publisher of specific image tags.

Through DCT, image publishers can sign their images and image consumers can ensure that the images they pull are signed. Publishers could be individuals or organizations manually signing their content or automated software supply chains signing content as part of their release process.

## Image tags and DCT

> [REGISTRY_HOST[:REGISTRY_PORT]/]REPOSITORY[:TAG]

- REPOSITORY: A particular image repository can have multiple tags. An image publisher can build an image and tag combination many times changing the image with each build
- TAG: DCT is associated with the TAG portion of an image. Each image repository has a set of keys that image publishers use to sign an image tag. Image publishers have discretion on which tags they sign
- An image repository can contain an image with one tag that is signed and another tag that is not. It is the responsibility of the image publisher to decide if an image tag is signed or not. Publishers can choose to sign a specific tag or not.

## Docker Content Trust Keys

Trust for an image tag is managed through the use of signing keys. A key set is created when an operation using DCT is first invoked. A key set consists of the following classes of keys:
- an offline key that is the root of DCT for an image tag
- repository or tagging keys that sign tags
- server-managed keys such as the timestamp key, which provides freshness security guarantees for your repository


## Signing Images with Docker Content Trust

### Prerequisites

- A Docker Registry with a Notary server attached

### Install notaly localy

```
sudo apt-get install notary
```

### Deploy Notary
```sh
git clone https://github.com/theupdateframework/notary.git
cd notary
docker-compose build
docker-compose up -d
mkdir -p ~/.notary && cp cmd/notary/config.json cmd/notary/root-ca.crt ~/.notary

# If you want to stop notary just run
# docker-compose down 
# And to remove the containers completely
# docker-compose down --rmi all
```

### Deploy local docker registry
```sh
docker run -d -p 5000:5000 --restart=always --name local_registry registry:2
```

### Create necessary environment variables
```sh
export DOCKER_CONTENT_TRUST_SERVER=https://localhost:4443
export DOCKER_CONTENT_TRUST=1
# this will disallow you from using non signed images
```

### Generate a new key
```sh
docker trust key generate <user>
# Generating key for jeff...
# Enter passphrase for new <user> key with ID 9deed25:
# Repeat passphrase for new <user> key with ID 9deed25:
# Successfully generated and loaded private key. Corresponding public key # available: /home/ubuntu/Documents/mytrustdir/<user>.pub
```

### Add signed to admin/repo
```sh
docker trust signer add --key <user>.pub user localhost:5000/admin/demo 
```

### Try to download and upload a signed image to local registry
```sh
# Pull and upload hello-world docker image to local registry
docker pull hello-world
docker tag hello-world localhost:5000/hello-world:v1.0.0
docker push localhost:5000/hello-world:v1.0.0

# Inspect the signed container
docker trust inspect --pretty localhost:5000/hello-world:v1.0.0
```

## Debugging
### Notary certificate problem

```sh
# Needs gcc, make, cfssl, cfssljson
sudo apt install golang-cfssl make gcc

cd notary
cd fixtures

# Change of file `regenerateTestingCerts.sh` line `174`
# From `echo >&2 "Installing cfssl tools"; go get -u github.com/cloudflare/cfssl/cmd/...;`
# To `echo >&2 "Installing cfssl tools"; go get github.com/cloudflare/cfssl/cmd/...;`
./regenerateTestingCerts.sh
```


## References

- https://docs.docker.com/engine/security/trust/
- https://docs.docker.com/engine/security/trust/deploying_notary/
- https://docs.docker.com/notary/running_a_service/
- https://stackoverflow.com/questions/48261747/how-do-we-setup-docker-notary-server-notary-signer-and-notary-client-for-priva
- https://marcofranssen.nl/signing-docker-images-using-docker-content-trust/
- [Certificate problem](https://github.com/theupdateframework/notary/issues/1593)