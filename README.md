[![Go Report Card](https://goreportcard.com/badge/github.com/matchstick/exifsort)](https://goreportcard.com/report/github.com/matchstick/exifsort)
[![CI](https://img.shields.io/badge/CI-passing-brightgreen)](https://github.com/matchstick/exifsort/actions)
![Coverage](https://img.shields.io/badge/coverage-93%25-brightgreen)
![golangci-lint](https://img.shields.io/badge/golangci--lint-100%25-brightgreen)
[![GoDoc](https://godoc.org/github.com/matchstick/exifsort?status.svg)](https://godoc.org/github.com/matchstick/exifsort)

# exifsort

Libraries and CLI to sort media by date using EXIF and file modTime information.

This is for folks who have a closet full of hard drives and network drives full
of photos and want to centralize them in one folder structure that is organized
by time.

exifsort has been tested on over 120K files taken from 1999 til present day. It is
also given special case for being run on a Synology to avoid special metadata
files on those servers.

exifsort will process these [file extensions](https://godoc.org/github.com/matchstick/exifsort/lib#pkg-constants) and skip other files.

# Overview

The library and API live in the lib directory. Check out
[exifsort/lib godocs](https://godoc.org/github.com/matchstick/exifsort/lib).

The program is written to employ several stages to let the user transform a set of
directories of random directories with unorganized photos to one centralized
sorted directory. Much care has been taken to ensure no files are removed that
are not duplicates among this series of stages. Along the way it tries to weed
out duplicates. This directory is sorted by one of three **methods**.

exifsort will try to use the file exif data's **IFD/EXIF/DateTimeOriginal** field
to determine how to sort the media. If it cannot find exif data it uses file modtime.

## Installation

To build exifsort you need to use standard Makefile.

`$ make`

Creates exifsort.

Cross compilation is also supported in the Makefile.

## Usage

If you have a directory full of files and photos called **random** and you want
to copy them to a new directory **sortedNew** sorted by month that only has photos.

Simple Example:

` $ exifsort sort copy month random/ sortedNew/`

JSON Example:

`$ exifsort scan random/ -j random.json`

`$ exifsort sort copy month random.json sortedNew/`

# Stages

exifsort is intended to be used in sequential stages.

| Stage | Description |
|-------|-------------|
| Scan  | Collect the mapping of file path to time it was created |
| Sort  | Use scan mapping to transfer files to newly created sorted directory organized by a method. |
| Merge | Transfer files from one sorted directory to another sorted directory. |

## Methods

Sorted directories are organized by a method.

| Method | Structure | Example |
| ------ | --------- | ------- |
| Year   | dst -> year -> media | dst/2020/pic.jpg |
| Month  | dst -> year-> month -> media | dst/2020/2020_04/pic.jpg |
| Day    | dst -> year-> month -> day -> media | dst/2020/2020_4/20202_04_12/pic.jpg |

# Commands

## scan

**exifsort scan** reads the data from the directory of files, and builds a mapping of path to time created.

Useful to test that exif library and program has no surprises.

Example:

`$ exifsort scan data/`

You can save data to a json file too:

`$ exifsort scan data/ -j src.json`

## sort

The sort command performs a number of steps. It can also optionally scan and sort in one command.

  1. Create a directory for output
  1. Collect file mapping to creating time via scanning a directory or reading a json file
  1. Indexing the media by method. To prevent filename collisions sort renames files.
  1. If using **move** it will remove duplicates in the src directory.
  1. Transfer media to the output


Examples:

`$ exifsort sort copy month src/ dst/`

Will create a new directory called dst, scan the media in src, index that media
then **copy** it to dst so that it is arranged by **month**.

`$ exifsort sort move year src/ dst/`

Will create a new directory called dst, scan the media in src, index that media
then **move** the files to dst so that it is arranged by **year**. 

Example of json input:

`$ exifsort sort copy month src.json dst/`

You don't want to modify the directory that you scanned to generate the json
file between generating it and then sorting. This allow you to only scan
once.

## merge

Merge output from a sorted directory to another sorted directory.

`$ exifsort merge copy src/ dst/ `

Will merge two sorted directories and **copy** files from one to the other.

`$ exifsort merge move src/ dst/ `

Will merge two sorted directories and **move** files from one to the other.

When using subcommand **merge** and **move** together exifsort removes duplicates in the
src directory.

## filter

Filter is identical to merge except it accepts a regular expression as an argument.
This is used to match the files in the src diretory. Only those that match are
then merged to the dst directory.

`$ exifsort filter src/ dst/ "regex"`

## eval

scans by file not directory. Prints the date information of files specified.

`$ exifsort eval data/*`

# Thanks

Huge thanks to [dsoprea](https://github.com/dsoprea) for his [exif library](https://github.com/dsoprea/go-exif)
