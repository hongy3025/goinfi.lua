#!/bin/sh
PWD=`pwd`
LIBPATH=$GOPATH/lib
export LD_LIBRARY_PATH=`realpath $LIBPATH`
exec $PWD/$1
