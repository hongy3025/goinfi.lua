GOARCH=$(shell go env GOARCH)
GOBIN=$(shell go env GOBIN)
GOCHAR=$(shell go env GOCHAR)
GOEXE=$(shell go env GOEXE)
GOGCCFLAGS=$(shell go env GOGCCFLAGS)
GOHOSTARCH=$(shell go env GOHOSTARCH)
GOHOSTOS=$(shell go env GOHOSTOS)
GOOS=$(shell go env GOOS)
GOPATH=$(shell go env GOPATH)
GOROOT=$(shell go env GOROOT)
GOTOOLDIR=$(shell go env GOTOOLDIR)
CGO_ENABLED=$(shell go env CGO_ENABLED)
PKG_ROOT=$(GOPATH)/pkg/$(GOOS)_$(GOARCH)

GO_SRC=convert.go lua.go pack.go rawlua.go
C_SRC=clua.c
MYCFLAGS=-I lua-5.1.5/src
MYFLAGS=-L ../../../lib -llua -lm -ldl 
PKG_NAME=lua
PKG_PATH=goinfo

CGOGO=$(addprefix _cgo/,$(subst .go,.cgo1.go,$(GO_SRC)))
CGO_SRC=$(addprefix _cgo/,$(subst .go,.cgo2.c,$(GO_SRC)))
ALL_SRC=$(GO_SRC) $(C_SRC) $(CGO_SRC)
CGO_STUB=_cgo/_cgo_main.o _cgo/_cgo_export.o
CGO_OBJ=$(CGO_SRC:.c=.o) $(CGO_STUB)
C_OBJ=$(subst .c,.o,$(C_SRC))
ALL_OBJ=$(CGO_OBJ) $(C_OBJ)
CGO_DEFUN=_cgo/_cgo_defun.6
CGO_IMPORT=_cgo/_cgo_import.6
GOBJ=_go_.6 $(CGO_IMPORT) $(CGO_DEFUN)
CGO=_cgo/cgogen
CFLAGS=$(GOGCCFLAGS) $(MYCFLAGS) -I. -I _cgo
LDFLAGS=$(MYFLAGS)
PKG_FNAME=$(PAG_NAME)/$(PKG_NAME)
TARGET=$(PKG_NAME).a

all:
	@gmake --no-print-directory prepare
	@gmake --no-print-directory pack

prepare:
	@[ -d _cgo ] || mkdir _cgo

pack: $(TARGET)

$(TARGET): $(CGO) $(GOBJ) $(ALL_OBJ)
	@echo hahaah
	go tool pack grc $(TARGET) $(GOBJ) $(ALL_OBJ)

_cgo_.o: $(ALL_OBJ)
	gcc $(CFLAGS) -o _cgo_.o $(LDFLAGS) $(ALL_OBJ)

$(CGO): $(GO_SRC)
	go tool cgo -objdir _cgo -- $(C_INCLUDE) $(GO_SRC)
	touch $(CGO)

_go_.6: $(CGO) _cgo/_cgo_gotypes.go $(CGOGO)
	go tool 6g -o _go_.6 -p $(PKG_FNAME) -I $(GOPATH)/pkg/$(GOOS)_$(GOARCH) -I $(GOROOT)/pkg/$(GOOS)_$(GOARCH) _cgo/_cgo_gotypes.go $(CGOGO)

$(CGO_IMPORT): _cgo/_cgo_import.c
	@echo abcdefg
	go tool 6c -FVw $(C_INCLUDE) -I $(GOROOT)/pkg/$(GOOS)_$(GOARCH) -o $(CGO_IMPORT) -DGOOS_$(GOOS) -DGOARCH_$(GOARCH) _cgo/_cgo_import.c

$(CGO_DEFUN): $(CGO)
	go tool 6c -FVw $(C_INCLUDE) -I $(GOROOT)/pkg/$(GOOS)_$(GOARCH) -o $(CGO_DEFUN) -DGOOS_$(GOOS) -DGOARCH_$(GOARCH) _cgo/_cgo_defun.c


_cgo/_cgo_import.c: _cgo_.o
	go tool cgo -objdir _cgo -dynimport _cgo_.o -dynout _cgo/_cgo_import.c

install:
	@[ -d $(PKG_ROOT)/$(PKG_PATH) ] | mkdir -p $(PKG_ROOT)/$(PKG_PATH)
	@cp -v $(TARGET) $(PKG_ROOT)/$(PKG_PATH)/$(TARGET)

clean:
	@rm -f $(TARGET)
	@rm -f _cgo_.o
	@rm -f *.6
	@rm -f $(ALL_OBJ)
	@rm -rf _cgo

echo:
	@echo src: $(ALL_SRC)
	@echo obj: $(ALL_OBJ)
	@echo $(CGOGO)

.c.o:
	gcc $(CFLAGS) -o $@ -c $<

.PHONY: prepare all echo cgo compile


