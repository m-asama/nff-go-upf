#
#PATH_TO_MK = ${HOME}/go/pkg/mod/github.com/intel-go/nff-go@v0.9.2/mk
PATH_TO_MK = ../nff-go/mk
include $(PATH_TO_MK)/leaf.mk

nff-go-upf:
	go build .
