# OSAS - Open Storage Account Scanner ;)

**First off, this is not a polished project. If you are expecting a turn-key solution, you're going to be very dissapointed.**
**Even if you are not expecting a turn-key solution, you'll probably still be dissapointed.**

This was hacked together for a personal project of mine to grab a random sampling of Azure storage accounts and see how many were wide open.
Similar to the various AWS S3 bucket scanners out there, this simply looks for any storage account blob containers that have been left open, and then stores the results in a database. There are also some utility tools to help dig through the data. 

Contents (each has a README):
* [lister](): a tool to list contents, and optionally download those contents, of an open blob container
* [scanner-gui](): a small tool to display the results as a webpage
* [scanner](): this is the actual application that does the scanning

You fire up however many `scanner` instances you want, and then you can use `lister` to display the contents of the open blob containers `scanner` finds.

As everything is stored in a database, the automation possibilities are pretty big. Maybe automate the calling of `lister` on open accounts and store those results in another table? Maybe take those results and, based on filetype, do some OCR? 
