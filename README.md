# AWS Terminal UI (aws-tui)

[![main](https://github.com/giovannirossini/aws-tui/actions/workflows/main.yml/badge.svg)](https://github.com/giovannirossini/aws-tui/actions/workflows/main.yml) [![release](https://github.com/giovannirossini/aws-tui/actions/workflows/release.yml/badge.svg)](https://github.com/giovannirossini/aws-tui/actions/workflows/release.yml)

AWS Terminal UI is a fast, terminal-first TUI written in Go that lets you browse and manage AWS resources without leaving your shell.

It provides an interactive terminal interface to explore AWS services, view resources, and access key information across multiple AWS profiles and regions.

## Why it exists

Managing AWS resources typically means switching between the AWS Console, CLI commands, and documentation. AWS Terminal UI removes that friction by bringing AWS resource management straight to your terminal with an intuitive, keyboard-driven interface.

## Key features

- Interactive terminal UI built with Bubble Tea
- Support for 25+ AWS services (EC2, S3, Lambda, RDS, VPC, and more)
- Multi-profile support with easy profile switching
- Fast resource browsing with caching
- Clean, organized service categories
- Real-time data fetching from AWS APIs
- Keyboard-first navigation

## Usage

```sh
aws-tui
```

The application will start with a profile selector, then display the main service menu. Navigate using arrow keys, select services, and explore your AWS resources.

## Installation

```sh
» make

Building aws-tui...
go build -ldflags "-s -w" -o bin/aws-tui cmd/aws-tui/main.go
```

Move the binary to your PATH:

```sh
sudo mv bin/aws-tui /usr/local/bin
```

## Design philosophy

- Zero configuration, works with your existing AWS credentials
- Fast and responsive terminal experience
- Clean separation of concerns and idiomatic Go code
- Lazy loading of AWS clients for optimal startup time

AWS Terminal UI is built for engineers who live in the terminal and want instant access to AWS resources without context switching.
