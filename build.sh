#!/bin/bash
rm -f nff-go-upf
NFF_GO_NO_MLX_DRIVERS=yes NFF_GO_NO_BPF_SUPPORT=yes make nff-go-upf
