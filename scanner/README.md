
This ~~POS~~ PoC can be used to scan Azure storage accounts for blob containers that have public read access enabled.
When you use `?restype=container&comp=list` in your request, a `200` will be returned if the container is
open, so this thing just looks for that [(API docs)](https://docs.microsoft.com/en-us/rest/api/storageservices/list-blobs).


**Now with moar power!!!!**

I decided to make this a bit more scalable, so now it uses a central database that "workers" can pull from to
do their thing. The idea is that you could now do something like shove this into a docker container (notice the `Dockerfile`?) and scan things with a ton of workers. Eventually I'll add some type of brute force thingamajig so it doesn't
need to use a wordlist.

You'll need a MySQL (or whatever they call it these days) database and two tables:

    CREATE TABLE stg_accts (
      url VARCHAR(2083) NOT NULL,
      last_check datetime NOT NULL,
      is_open tinyint(1)  NOT NULL
    );


    CREATE TABLE open_containers (
      url VARCHAR(2083) NOT NULL,
      last_check datetime NOT NULL
    );


The `stg_accts` table is going to be a list of URLs you want to scan. You are in charge of getting those URLs into the database. When the app starts, it'll
hit that table, find the oldest record and start scanning it. If it finds that it is open, it will update the
table PLUS put whatever results it finds into the `open_containers` table.

There is a sample `config.json` that shows you what your config should look like.

This is a really silly app that was hacked together quickly for fun. Please don't expect it to actually work.


    Usage of ./scanner:
      -c string
    	    config file location (default "conf.json")
      -m int
    	    max connections to host (default 100)
      -w int
    	    worker count (default 10)
