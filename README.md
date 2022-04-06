# OSAS - Open Storage Account Scanner ;)

**First off, this is not a polished project. If you are expecting a turn-key solution, you're going to be very disappointed.**
**Even if you are not expecting a turn-key solution, you'll probably still be disappointed.**

This was hacked together for a personal project of mine to grab a random sampling of Azure storage accounts and see how many were wide open.
Similar to the various AWS S3 bucket scanners out there, this simply looks for any storage account blob containers that have been left open, and then stores the results in a database. There are also some utility tools to help dig through the data. 


OSAS has been re-factored to ship as an all-in-one tool vs multiple specific tools, hopefully making it easier to use. In the current form there
are three supported commands:

* list: list contents, and optionally download those contents, of an open blob container
* load: load a list of targets into the database
* scan: this is the actual application that does the scanning

You fire up however many `scan` instances you want, and then you can use `list` to display the contents of the open blob containers `scan` finds.

As everything is stored in a database, the automation possibilities are pretty big. Maybe automate the calling of `list` on open accounts and store those results in another table? Maybe take those results and, based on file type, do some OCR? 


# Install & Setup

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


The `stg_accts` table is going to be a list of URLs you want to scan, this is where the `load` command will insert your targets. When the app starts, it'll
hit that table, find the oldest record and start scanning it. If it finds that it is open, it will update the
table PLUS put whatever results it finds into the `open_containers` table.

There is a sample `config.json` that shows you what your config should look like.


# Usage

Using this is pretty straightforward and doing an `OSAS -h` is probably the best way to get going. One thing to keep in mind is that `scan` is reading in a 
list of targets from the database, so this is meant as more of a bulk scanning type tool. Feel free to fire up multiple instances of `scan` to work through
your target list more quickly.
