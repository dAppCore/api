package miner

import newminer "dappco.re/go/miner"

const MinerTypeXMRig = newminer.MinerTypeXMRig

type AvailableMiner = newminer.AvailableMiner
type ProviderRouteHandler = newminer.ProviderRouteHandler
type Service = newminer.Service

var DecodeProviderConfig = newminer.DecodeProviderConfig
var DecodeProviderProfile = newminer.DecodeProviderProfile
var DecodeProviderStdin = newminer.DecodeProviderStdin
var DefaultMetrics = newminer.DefaultMetrics
var ErrProfileNotFound = newminer.ErrProfileNotFound
var ErrServiceNotReady = newminer.ErrServiceNotReady
var NewService = newminer.NewService
