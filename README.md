# exifsort

![Under Construction](data/construction.jpg) 

Program are both under construction but for now here are some notes.

# exifsort

Libraries and CLI to sort to sort media by date using EXIF information.

This is for folks who have a closet full of hard drives and network drives full
of photos and want to centralize them in one folder structure.

The functionality and API live in the lib directory. Check out
[godocs](https://godoc.org/github.com/matchstick/exifsort/lib) for details.

# Overview

The program is written to employ several stages to let the user verify the
step results as they organize their photos. It cannot hurt to be careful.

Huge thanks to [dsoprea](https://github.com/dsoprea) for his [exif
library](https://github.com/dsoprea/go-exif) and fast responses.

exifsort will try to use the exifdata to determine the time period to sort the
media. If it cannot find one due to an error in the exif data it will then rely
on file modtime.

# Commands

## scan

Scanning is when exifsort will read the data from the directory of files,
filter for media and retrieve time. Useful to test that exif library will be fine.

`exifsort scan input -i <src> [--summarize --quiet]`

You can save data to a json file too:

`exifsort scan input -i <src> -j <json file> [--summarize --quiet]`

## sort

Walk an input directory, index the data and then transfer files to an output  directory.

`exifsort sort -i <src> -o <dst> -m <year | month | day> -a <copy | move>`

or load from json file:

`exifsort sort -j <json> -o <dst> -m <year | month | day> -a <copy | move>`

Sort supports -s and -q also.

### Methods

| Method | Structure |
| ------ | --------- |
| Year   | dst -> year -> media |
| Month  | dst -> year-> month -> media |
| Day    | dst -> year-> month -> day -> media |

### Actions

An action specifies whether to move or copy the files from input to output 

| Action | Transfer By |
| ------ | --------- |
| Copy   | copy file from src to dst |
| Move   | move file from src to dst |


## eval

scans by file not directory. Prints the date information of files specified. 

`exifsort eval <files>`

## merge

Soon we will be able to merge output from sort to a pre-existing directory.

`exifsort merge -i <src> -o <dst> -q -s`

TODOs
=====
* concurrency
* Write a merge step to merge two sorted directories.
	* check input and output directories are correct format
	* copy or move files from input to output
* Set up CI on github.
