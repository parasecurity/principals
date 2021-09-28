#!/bin/zsh

root_dir=`realpath $0 | sed 's/testing\/headers.sh//'`

function validate(){
	$root_dir/testing/validate.sh $@
}	

