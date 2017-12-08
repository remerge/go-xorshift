PROJECT = go-xorshift
PACKAGE = github.com/remerge/$(PROJECT)

GOMETALINTER_OPTS = --enable-all --tests -D unparam

include Makefile.common
