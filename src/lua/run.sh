#!/bin/sh
PWD=`pwd`
LIBPATH=$PWD/../../lib
export LD_LIBRARY_PATH=`realpath $LIBPATH`
exec $PWD/$1
