# `dmsgexec`

`dmsgexec` allows remote execution of commands via `dmsg`.

## Install

**Golang**

Ensure the latest version of Golang is installed. The project depends on go modules for dependency management.
Detailed golang installation steps can be found here: https://github.com/SkycoinProject/skycoin/blob/develop/INSTALLATION.md

**`dmsgexec-server` and `dmsgexec`**

The following commands will install the binaries into `$HOME/go/bin`. You can change the destination folder by changing the value of the `GOBIN` env.

Ensure the path you set as `GOBIN` is part of your `PATH` env.

```
# clone repository
git clone https://github.com/SkycoinProject/dmsgexec.git

# cd into repo
cd dmsgexec

# install all binaries into $HOME/bin
GOBIN=$HOME/go/bin go install ./...

```

### Run

`dmsgexec` is made up of two executables:
- `dmsgexec-server` is to be the service for dmsg exec.
- `dmsgexec` is the command line tool for interacting with `dmsgexec-server`.

The following instructions run `dmsgexec-server` with the default configurations.

For more advanced control, run `dmsgexec-server --help` or `dmsgserver --help` for additional arguments.

**Generate your local keys:**
```
$ dmsgexec keygen
```

**Run `dmsgexec-server`:**
```
$ dmsgexec-server
```

**Add remote public keys to your whitelist:**

Add trusted public keys to allow the associated remote `dmsgexec-server`s to execute commands on your local machine.
```
$ dmsgexec whitelist-add --pk <public-key>
```

**Execuate a command on a remote machine:**

```
$ dmsgexec --pk <public-key> --cmd "echo" --arg "hello world!"
$ dmsgexec --pk <public-key> --cmd "tree" --arg "-L" --arg "1"
```
