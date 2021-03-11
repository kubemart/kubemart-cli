#!/bin/bash

mkdir ~/kind
kind create cluster --name cluster-1 --kubeconfig ~/kind/cluster-1.yaml
kind create cluster --name cluster-2 --kubeconfig ~/kind/cluster-2.yaml
kind create cluster --name default
