SRC := forward.py
BUILD ?= .

.PHONY: all

ifneq ($(BUILD), .)
all: $(SRC)
	cp $(SRC) $(BUILD)/
else
all:
endif
