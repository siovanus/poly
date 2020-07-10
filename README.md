
<h1 align="center">Ontology </h1>
<h4 align="center">Version 1.0 </h4>

[![GoDoc](https://godoc.org/github.com/polynetwork/ontology?status.svg)](https://godoc.org/github.com/polynetwork/ontology)
[![Go Report Card](https://goreportcard.com/badge/github.com/polynetwork/ontology)](https://goreportcard.com/report/github.com/polynetwork/ontology)
[![Travis](https://travis-ci.org/polynetwork/ontology.svg?branch=master)](https://travis-ci.org/polynetwork/ontology)
[![Discord](https://img.shields.io/discord/102860784329052160.svg)](https://discord.gg/gDkuCAq)

English | [中文](README_CN.md)

Welcome to the official source code repository for Ontology!

Ontology is dedicated to developing a high-performance blockchain infrastructure, which is customizable to different business requirements. 

Prerequisites for getting started with development on the Ontology networks are:

- Mainstream coding and development experience
- Understanding of your business scenario/requirements
- No need for previous blockchain engineering experience

The Ontology core tech team, the community, and the ecosystem can all support you in development. MainNet, TestNet, Docker image for Ontology, SmartX, and Ontology Explorer combined make it easy to start.

Ontology makes getting started easier!

Ontology MainNet has been launched on Jun 30, 2018. <br>
And new features are still rapidly under development. The master branch may be unstable, but stable versions can be found under the [release page](https://github.com/polynetwork/poly/releases).

We openly welcome developers to Ontology.

## Features 

- Scalable lightweight universal smart contract
- Scalable WASM contract support
- Crosschain interactive protocol (processing)
- Multiple encryption algorithm support
- Highly optimized transaction processing speed
- P2P link layer encryption (optional module)
- Multiple consensus algorithm support (VBFT/DBFT/RBFT/SBFT/PoW)
- Quick block generation time


## Contents

- [Build development environment](#build-development-environment)
- [Get Ontology](#get-ontology)
    - [Get from release](#get-from-release)
    - [Get from source code](#get-from-source-code)
- [Run Ontology](#run-ontology)
    - [MainNet sync node](#mainnet-sync-node)
    - [Public test network Polaris sync node](#public-test-network-polaris-sync-node)
    - [Testmode](#testmode)
    - [Run in docker](#run-in-docker)
- [Some examples](#some-example)
    - [ONT transfer sample](#ont-transfer-sample)
    - [Query transfer status sample](#query-transfer-status-sample)
    - [Query account balance sample](#query-account-balance-sample)
- [Contributions](#contributions)
- [Open source community](#open-source-community)
    - [Site](#site)
    - [Developer Discord Group](#developer-discord-group)
- [License](#license)

## Build development environment
The requirements to build Ontology are:

- Golang version 1.9 or later
- Glide (a third party package management tool)
- Properly configured Go language environment
- Golang supported operating system

## Get Ontology

### Get from release
- You can download the latest Ontology binary file with ` curl https://dev.ont.io/ontology_install | sh `.

- You can download other versions at [release page](https://github.com/polynetwork/poly/releases).

### Get from source code

Clone the Ontology repository into the appropriate $GOPATH/src/github.com/polynetwork directory.

```
$ git clone https://github.com/polynetwork/ontology.git
```
or
```
$ go get github.com/polynetwork/ontology
```
Fetch the dependent third party packages with glide.

```
$ cd $GOPATH/src/github.com/polynetwork/ontology
$ glide install
```

If necessary, update dependent third party packages with glide.

```
$ cd $GOPATH/src/github.com/polynetwork/ontology
$ glide update
```

Build the source code with make.

```
$ make all
```

After building the source code sucessfully, you should see two executable programs:

- `ontology`: the node program/command line program for node control.
- `tools/sigsvr`: (optional) Ontology Signature Server - sigsvr is a RPC server for signing transactions for some special requirements. Detailed docs can be found [here](https://github.com/polynetwork/documentation/blob/master/docs/pages/doc_en/Ontology/sigsvr_en.md).

## Run Ontology

You can run Ontology in four different modes:

1) MainNet (./ontology)
2) TestNet (./ontology --networkid 2)
3) Testmode (./ontology --testmode)
4) Docker

E.g. for Windows (64-bit), use command prompt and cd to the dirctory where you installed the Ontology release, then type `start ontology-windows-amd64.exe --networkid 2`. This will sync to TestNet and you can explore further by the help command `ontology-windows-amd64.exe --networkid 2 help`.

### MainNet sync node

Run ontology directly

   ```
	./ontology
   ```
then you can connect to Ontology MainNet.

### Public test network Polaris sync node (TestNet)

Run ontology directly

   ```
	./ontology --networkid 2
   ```
   
Then you can connect to the Ontology TestNet.



### Run in docker

Please ensure there is a docker environment in your machine.

1. Make docker image

    - In the root directory of source code, run `make docker`, it will make an Ontology image in docker.

2. Run Ontology image

    - Use command `docker run polynetwork/ontology` to run Ontology；

    - If you need to allow interactive keyboard input while the image is running, you can use the `docker run -ti polynetwork/ontology` command to start the image;

    - If you need to keep the data generated by image at runtime, you can refer to the data persistence function of docker (e.g. volume);

    - If you need to add Ontology parameters, you can add them directly after `docker run polynetwork/ontology` such as `docker run polynetwork/ontology --networkid 2`.
     The parameters of ontology command line refer to [here](./docs/specifications/cli_user_guide.md).

## Some examples

### ONT transfer sample
 -- from: transfer from； -- to: transfer to； -- amount: ONT amount；
```shell
 ./ontology asset transfer  --from=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 --to=AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce --amount=10
```
If the asset transfer is successful, the result will display as follows:

```shell
Transfer ONT
  From:ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
  To:AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce
  Amount:10
  TxHash:437bff5dee9a1894ad421d55b8c70a2b7f34c574de0225046531e32faa1f94ce
```
TxHash is the transfer transaction hash, and we can query a transfer result by the TxHash.
Due to block time, the transfer transaction will not be executed before the block is generated and added.

If you want to transfer ONG, just add --asset=ong flag.

Note that ONT is an integer and has no decimals, whereas ONG has 9 decimals. For detailed info please read [Everything you need to know about ONG](https://medium.com/ontologynetwork/everything-you-need-to-know-about-ong-582ed216b870).

```shell
./ontology asset transfer --from=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 --to=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 --amount=95.479777254 --asset=ong
```
If transfer of the asset succeeds, the result will display as follows:

```shell
Transfer ONG
  From:ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
  To:AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce
  Amount:95.479777254
  TxHash:e4245d83607e6644c360b6007045017b5c5d89d9f0f5a9c3b37801018f789cc3
```

Please note, when you use the address of an account, you can use the index or label of the account instead. Index is the sequence number of a particular account in the wallet. The index starts from 1, and the label is the unique alias of an account in the wallet.

```shell
./ontology asset transfer --from=1 --to=2 --amount=10
```

### Query transfer status sample

```shell
./ontology info status <TxHash>
```


For further examples, please refer to the [CLI User Guide](https://polynetwork.github.io/documentation/cli_user_guide_en.html).

## Contributions

Please open a pull request with a signed commit. We appreciate your help! You can also send your code as email to the developer mailing list. You're welcome to join the Ontology mailing list or developer forum.

Please provide a detailed submission information when you want to contribute code for this project. The format is as follows:

Header line: Explain the commit in one line (use the imperative).

Body of commit message is a few lines of text, explaining things in more detail, possibly giving some background about the issue being fixed, etc.

The body of the commit message can be several paragraphs. Please do proper word-wrap and keep columns shorter than 74 characters or so. That way "git log" will show things  nicely even when it is indented.

Make sure you explain your solution and why you are doing what you are doing, as opposed to describing what you are doing. Reviewers and your future self can read the patch, but might not understand why a particular solution was implemented.

Reported-by: whoever-reported-it +
Signed-off-by: Your Name [youremail@yourhost.com](mailto:youremail@yourhost.com)

## Open source community
### Site

- <https://ont.io/>

### Developer Discord Group

- <https://discord.gg/4TQujHj/>

## License

The Ontology library is licensed under the GNU Lesser General Public License v3.0, read the LICENSE file in the root directory of the project for details.
