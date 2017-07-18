## Manifold CLI

The Manifold CLI tool makes is easy to create and list resources, export
credentials, and inject configuration directly into a running application.

If you have any feedback, ideas, or ways we could improve the tool please don't
hesitate to reach out to us via email at [ian@manifold.co](mailto:ian@manifold.co) or
[peter@manifold.co](mailto:peter@manifold.co).

## Getting Started

To get started, you must first install the CLI locally, a zip file for your
operating system of choice should have been included with these instructions.

To install, unzip the `manifold-cli` binary. You can install it globally by
placing it within your `$PATH` or run it locally from the unzip'd directory.

Once complete, you should be able to run the `manifold-cli` command! To login
to your Manifold account, use the `manifold-cli login` command.

```bash
$ manifold-cli login
✔ Email: ian@manifold.co
✔ Password: ●●●●●●●●
You are logged in, hooray!
```

**NOTE**: All examples assume that command line tool exists within your `$PATH`.

### Creating a Resource

You can provision a resource using the `manifold-cli create` which supports
many different flags and arguments for scripting purposes. It also offers an
interactive wizard for walking through the entire process!

```bash
$ manifold-cli create
✔ Select Product: LogDNA (logdna)
✔ Select Plan: Quaco (quaco) - Free!
✔ Region: All Regions (all::global)
✔ App Name: my amazing app
✔ Resource Name (one will be generated if left blank): logging-production

We're starting to create an instance of LogDNA. This may take some time, please wait!

An instance named "logging-production" has been created!
```

**TIP**: You can list out the options for any commands using the `-h` flag
(e.g. `manifold-cli create -h`).

### Listing Resources

You can list out all resources and their status using the `manifold-cli list`
command. Any resource which is ready to be used by your application will be
listed as `Ready`.

```bash
$ manifold-cli list
RESOURCE NAME                                     APP NAME              STATUS          PRODUCT                        PLAN                REGION

Efficacious Misty Moss Square                                           Ready           Bonsai Elasticsearch           Sandbox             AWS - US East 1 (N. Virginia)
Brown Cinnamon Satin Heptagon                                           Ready           LogDNA                         Quaco               All Regions
Bizarre Coffee Pentadecagon                       test                  Ready           LogDNA                         Quaco               All Regions
Intentional Dark Lava Octadecagon                 test                  Ready           CloudAMQP                      Little Lemur        AWS - US West 1 (N. California)
Anchored Deep Carmine Dihedron                    test                  Ready           LogDNA                         Quaco               All Regions
```

**TIP**: You can limit any output to a specific application using the `-a` flag
(e.g. `manifold-cli list -a test`).

### Exporting Credentials

You can export credentials from your provision resources using the
`manifold-cli export` command which will output into various different formats
using the `-f` flag.

```bash
$ manifold-cli export -a test
# Bizarre Coffee Pentadecagon
ACCOUNT=3b6f5sdff
KEY=dd6bfbd68a14474283b9502dsdfsdf
RESOURCE_ID=2683f9j8u67n9g2f3r74sdfsdf

# Intentional Dark Lava Octadecagon
CLOUDAMQP_URL=amqp://pvoksdfsdf:sdfsdfU6WwzxYE6KZckrUEjb3A@donkey.rmq.cloudamqp.com/pvoknoib

# Anchored Deep Carmine Dihedron
ACCOUNT=3b6f5bc6sdfsdf
KEY=dd6bfbd68a14474283b9502d8e6sdfsdf
RESOURCE_ID=268eentnn0x4fx5vp5947404sdfsdf
```

**Example: Exporting Credentials into the Current Shell**

```bash
$ eval $(manifold-cli export -a test -f bash)
```

**Example: Exporting Credentials into a .env file**

```bash
$ manifold-cli export -a test > .env
```

### Injecting Credentials

You can inject credentials directly into an application using the `manifold-cli
run` command which exposes them to a process as environment variables.

```bash
$ manifold-cli run -a test -- bin/www
```

You can combine these with your account login details to login and run a
process in one command using environment variables.

**Example: Injecting Credentials and Authenticating in One Line**

```bash
$ MANIFOLD_EMAIL=me@test.com MANIFOLD_PASSWORD=pencil manifold-cli run -- bin/www
```

**Example: Injecting Credentials and Authenticating by exporting variables**

```bash
$ export MANIFOLD_EMAIL=me@test.com
$ export MANIFOLD_PASSWORD=pencil
$ manifold-cli run -- bin/www
```

**NOTE**: Any `manifold-cli` command will attempt to login using the set
environment variables if you're not already authenticated!
