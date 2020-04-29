# exifSort
Program to sort photos and movies via Exif

Inputs:
* input media to sort directory (unsorted)
* output directory - sorted to year => month
* Unsorted directory
* move vs copy 
* Progress mode
* overrride log file exifSort.og

Take in input:
	* Config file? 
	* arg lines?

exifSort 
 provide help and list

exifSort -> scan <in dir> <#cpus> <log file> <progress>
 * walk a directory report of exif state and number of files

exifSort -> sort <in dir> <out dir>  <#cpus> <log file>
 * walk a directory and then process files to a target directory of nested 

exifSort -> eval <file>
 * Just extract one file's time info
	* Totals, percentage
	* Exif found
	* Exif not found
