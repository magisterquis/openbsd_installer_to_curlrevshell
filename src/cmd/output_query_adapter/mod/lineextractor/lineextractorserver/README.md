lineextractorserver
===================
Test server for lineextractor

Quickstart
----------
1.  Run the thing, note the listen address
```sh
go run .
```
2.  Send strange queries
```sh
    ftp -o- -V -M  'http://127.0.0.1:4444/foo?bar???=tridge baaz#quux//./.../.././././../../../..'
```


Usage
-----
```
Usage: lineextractorserver [options]

Test server for lineextractor.

Accepts HTTP requests and prints the lines extracted by lineextractor.

Terminates when stdin is closed.

Options:
  -listen address
    	Listen address (default "127.0.0.1:0")
  -print-quoted
    	Quote lines before printing
  -print-request-uri
    	Print the raw request URI
```
