package main

const Header = `<?xml version="v1.0" encoding="UTF-8"?>` + "\n"

type Message struct {
	BeginFlag  int16
	SendSyn    int64
	ReceievSyn int64
	SourceFlag int
	DataLen    int32
	Data       []byte
	EndFlag    int16
}

type Peoples struct {
	Nums  int32 `xml:"Nums,attr"`
	Items Items `xml:"Items"`
}

type Items struct {
	Items []Item `xml:"Item"`
}

type Item struct {
	Name string `xml:"Name,attr"`
}
