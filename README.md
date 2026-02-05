OpenBSD Installer to [Curlrevshell](https://github.com/magisterquis/curlrevshell) Adapter
=========================================================================================
Little adapter to get a [curlrevshell](
https://github.com/magisterquis/curlrevshell) shell from the OpenBSD installer
using [`ftp(1)`](https://man.openbsd.org/ftp.1) instead of curl.

Meant for debugging an install when neither graphical output nor a serial
console are available.

The general idea is to use this repository to generate a `miniroot.img` which
brings up networking and calls back with a shell.  The shell's input comes from
a single HTTPS stream to curlrevshell and each line of the shell's output will
be sent in its own HTTPS request via [`output_query_adapter`](
./src/cmd/output_query_adapter) to curlrevshell.

Quickstart
----------
1.  Clone this repository somewhere running OpenBSD:
    ```sh
    git clone https://github.com/magisterquis/openbsd_installer_to_curlrevshell.git
    cd openbsd_installer_to_curlrevshell
    ```
2.  Edit [`config.mk`](./config.mk) to have at least the correct address
    for curlrevshell, but also whatever other bits need changing.
    Variables can also be given to make at build time:
    ```sh
    vi config.mk
    ```
3.  Build a miniroot image:
    ```sh
    make # Uses doas, sorry :|
    ```
4.  In one shell, start [`output_query_adapter`](./src/cmd/output_query_adapter):
    ```sh
    ./start.sh output_query_adapter
    ```
5.  In another shell, start [`curlrevshell`](
    https://github.com/magisterquis/curlrevshell):
    ```sh
    ./start.sh curlrevshell
    ```
6.  Boot the miniroot image:
    ```sh
    # This one's very situationally-dependent, but could be something like
    ssh root@test 'cat >/dev/sda && sync && echo b >/proc/sysrq-trigger' <miniroot_amd64_crs.img
    ```
7.  Wait a bit for networking to come up and a shell to call back.

If all went well,
the installer, if visible, should look like
```
Welcome to the OpenBSD/arm64 7.8 installation program.
Starting non-interactive mode in 5 seconds...
(I)nstall, (U)pgrade, (A)utoinstall or (S)hell? 
Performing non-interactive install...
Terminal type? [vt220] vt220
System hostname? (short form, e.g. 'foo') crs-installer

Available network interfaces are: vio0 vlan0.
Network interface to configure? (name, lladdr, '?', or 'done') [vio0] vio0
IPv4 address for vio0? (or 'autoconf' or 'none') [autoconf] autoconf
IPv6 address for vio0? (or 'autoconf' or 'none') [none] none
Available network interfaces are: vio0 vlan0.
Network interface to configure? (name, lladdr, '?', or 'done') [done] done
Using DNS domainname my.domain
Using DNS nameservers at 10.0.0.1

Password for root account? 
Question has no answer in response file: "password for root account?"
1        ___________________
2       < In the installer! >
3        -------------------
4               \   ^__^
5                \  (oo)\_______
6                   (__)\       )\/\
7                       ||----w |
8                       ||     ||
failed; check /tmp/ai/ai.log
(I)nstall, (U)pgrade, (A)utoinstall or (S)hell? 
```
The `failed; check /tmp/ai/ai.log` there is expected.

`output_query_adapter` should look like
```
$ ./start.sh output_query_adapter
+ ./output_query_adapter -curlrevshell https://10.0.0.10:4444/o -tls crs.txtar
2026/02/05 22:10:22 Serving HTTPS on 0.0.0.0:5555
2026/02/05 22:11:17 [10.0.0.20:32770] Opened new connection for 1cd74e2x1pr3t
```

`curlrevshell` should look like
```
$ ./start.sh curlrevshell
+ go run -trimpath -ldflags -w -s github.com/magisterquis/curlrevshell@latest -callback-address 10.0.0.10:4444 -template crs.tmpl -tls-certificate-cache crs.txtar
22:10:24.046 Welcome to curlrevshell version v0.0.1-beta.8
22:10:24.047 Listening on 0.0.0.0:4444
22:10:24.048 To get a shell:

ftp -M -o- -S cafile=/etc/ssl/crs_cert.pem -V -w 15 https://10.0.0.10:4444/c | /bin/sh

22:11:17.164 [10.0.0.20] Sent script: ID:1cd74e2x1pr3t C2Addr:10.0.0.10:4444 Path:/c
22:11:17.584 [10.0.0.20] Input connected: ID "1cd74e2x1pr3t"
22:11:17.601 [10.0.0.10] Output connected: ID "1cd74e2x1pr3t"
22:11:17.601 [10.0.0.10] Shell is ready to go!
 ___________________
< In the installer! >
 -------------------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
>
```
and we have a shell :)

Theory
------
[`ftp(1)`](https://man.openbsd.org/ftp.1) seems to be the only thing on the
OpenBSD installer which speaks TLS, but it doesn't allow for the sort of 
long-lived connection to send shell output back to curlrevshell.  It does,
however, have a pretty generous idea of URL-encoding paths which can be used
to send lines of text back to curlrevshell via [`output_query_adapter`](
./src/output_query_adapter), which translates per-line requests into a
persistent stream.

Knobs
-----
The following files are handy for changing how things work.

### [`config.mk`](./config.mk)
Main configuration for building and callbacks.  Configures...
- Callback addresses
- TLS certificate common name
- Miniroot build things

### [`auto_install.conf`](./auto_install.conf)
Provides answers for installer questions, at least as far as necessary to
get networking up and running.

See [`autoinstall(8)`](https://man.openbsd.org/autoinstall.8) for more
information.

### [`crs.tmpl`](./crs.tmpl)
Curlrevshell's [`-template` template](
https://github.com/magisterquis/curlrevshell/blob/master/doc/template.md),
built along with the miniroot image.
