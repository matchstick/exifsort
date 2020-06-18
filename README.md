# exifsort

![Under Construction](data/construction.jpg) 

Everything is done except merge for 1.0 release.  Code coverage is greater than 93% and climbing.
Open [issues](https://github.com/matchstick/exifsort/issues) show remaining work.

# exifsort

Libraries and CLI to sort to sort media by date using EXIF information.

This is for folks who have a closet full of hard drives and network drives full
of photos and want to centralize them in one folder structure that is organized
by time.

# Overview

The library and API live in the lib directory. Check out
[exifsort/lib godocs](https://godoc.org/github.com/matchstick/exifsort/lib).


The program is written to employ several stages to let the user verify the
step results as they organize their photos. We break down the pipeline into
four stages so we can verify correctness for each stage.

| Stage | Description |
|-------|-------------|
| Scan  | Collect the mapping of path to time in json |
| Sort  | Accept scan results to transfer media to new directory organized by time. |
| Merge | Transfer one sorted directory to another sorted directory. |

## Actions

An action specifies whether to move or copy the files from input to output 

| Action | Transfer By |
| ------ | --------- |
| Copy   | copy file from src to dst |
| Move   | move file from src to dst |


## Methods

| Method | Structure |
| ------ | --------- |
| Year   | dst -> year -> media |
| Month  | dst -> year-> month -> media |
| Day    | dst -> year-> month -> day -> media |

exifsort will try to use the exifdata to determine the time period to sort the
media. If it cannot find one due to an error in the exif data it will then rely
on file modtime.

# Commands

## scan

Scanning is when exifsort will read the data from the directory of files,
filter for media and retrieve time. Useful to test that exif library will be
fine. You can optionally store the results in a json file.

Example:

`$ exifsort scan data/`

You can save data to a json file too:

`$ exifsort scan data/ -j data.json`

## sort

The sort command performs a number of steps:

  1. Collect media information via scanning a directory or reading a json file from scan
  1. Indexing the media by method
  1. Create a directory for output
  1. Transfer media to the output

Examples: 

`$ exifsort sort copy month src/ dst/`

Will create a new directory called dst, scan the media in src, index that media
then **copy** it to dst so that it is arranged by **month**.

`$ exifsort sort move year src/ dst/`

Will create a new directory called dst, scan the media in src, index that media
then **move** the files to dst so that it is arranged by **year**. 

Note: src can be either a directory or a json file.

## merge

Merge output from a sorted directory to another sorted directory.

`$ exifsort merge src/ dst/ <method>`

## filter

Filter is identical to merge except it accepts a regular expression as an argument. 
This is used to match the files in the src diretory. Only those are then merged to the dst directory.

`$ exifsort merge src/ dst/ <method> "regex"`

## eval

scans by file not directory. Prints the date information of files specified. 

`$ exifsort eval data/*`

# Thanks

Huge thanks to [dsoprea](https://github.com/dsoprea) for his [exif
library](https://github.com/dsoprea/go-exif) and fast responses.
