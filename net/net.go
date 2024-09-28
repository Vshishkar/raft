package net

import "net/rpc"

type PeerConnection struct {
	client *rpc.Client
}

func (pc *PeerConnection) Call(method string, args *interface{}, reply *interface{}) error {
	err := pc.client.Call(method, args, reply)
	return err
}
