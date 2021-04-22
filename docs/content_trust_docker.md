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

- A Docker Registry with a Notary server attached. Instructions for installation [here](https://docs.docker.com/engine/security/trust/deploying_notary/).

### Deploy local docker registry
```sh
docker run -d -p 5000:5000 --restart=always --name local_registry registry:2
```

### Deploy Notary
```
git clone https://github.com/theupdateframework/notary.git
cd notary
docker-compose build
docker-compose up -d
mkdir -p ~/.notary && cp cmd/notary/config.json cmd/notary/root-ca.crt ~/.notary
```


Within the Docker CLI we can sign and push a container image with the `$ docker trust`.
This is built on top of the Notary feature set.


## References

- https://docs.docker.com/engine/security/trust/
- https://docs.docker.com/notary/running_a_service/
- https://stackoverflow.com/questions/48261747/how-do-we-setup-docker-notary-server-notary-signer-and-notary-client-for-priva