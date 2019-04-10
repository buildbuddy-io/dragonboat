// Copyright 2017-2019 Lei Ni (nilei81@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"math"

	"github.com/lni/dragonboat/internal/utils/random"
)

const (
	// NotSessionManagedClientID is a special client id used to indicate
	// that the entry is not managed by a client session. This is used when
	// the entry is not generated by the user, e.g. the blank entry proposed
	// after becoming the leader.
	NotSessionManagedClientID uint64 = 0
	// NoOPSeriesID is a special series ID used to indicate that the client
	// session is a NOOP client session.
	NoOPSeriesID uint64 = 0
	// SeriesIDForRegister is the special series id used for registering
	// a new client session.
	SeriesIDForRegister uint64 = math.MaxUint64 - 1
	// SeriesIDForUnregister is the special series id used for unregistering
	// client session.
	SeriesIDForUnregister uint64 = math.MaxUint64
	// SeriesIDFirstProposal is the first series id to be used for making
	// proposals.
	SeriesIDFirstProposal uint64 = 1
)

// NewSession returns a new client session not registered yet. This function
// is not expected to be directly invoked by application.
func NewSession(clusterID uint64, rng random.Source) *Session {
	for {
		cid := rng.Uint64()
		if cid != NotSessionManagedClientID {
			return &Session{
				ClusterID: clusterID,
				ClientID:  cid,
				SeriesID:  NoOPSeriesID + 1,
			}
		}
	}
}

// NewNoOPSession creates a new NoOP client session ready to be used for
// making proposals. This function is not expected to be directly invoked by
// application.
func NewNoOPSession(clusterID uint64, rng random.Source) *Session {
	for {
		cid := rng.Uint64()
		if cid != NotSessionManagedClientID {
			return &Session{
				ClusterID: clusterID,
				ClientID:  cid,
				SeriesID:  NoOPSeriesID,
			}
		}
	}
}

// IsNoOPSession returns a boolean flag indicating whether the session instance
// is a NoOP session.
func (cs *Session) IsNoOPSession() bool {
	return cs.SeriesID == NoOPSeriesID
}

// ClusterIDMustMatch asserts that the input cluster id matches the cluster id
// of the client session.
func (cs *Session) ClusterIDMustMatch(clusterID uint64) {
	if cs.ClusterID != clusterID {
		panic("cluster id do not match")
	}
}

// PrepareForRegister sets the series id to the special series id for
// registering client session.
func (cs *Session) PrepareForRegister() {
	cs.assertRegularSession()
	cs.SeriesID = SeriesIDForRegister
}

// PrepareForUnregister sets the series id to the special series id for
// unregistering client session.
func (cs *Session) PrepareForUnregister() {
	cs.assertRegularSession()
	cs.SeriesID = SeriesIDForUnregister
}

// PrepareForPropose sets the series id to the first series id that can be used
// for making proposals.
func (cs *Session) PrepareForPropose() {
	cs.assertRegularSession()
	cs.SeriesID = SeriesIDFirstProposal
}

// ProposalCompleted increases the series id and the RespondedTo value.
// ProposalCompleted is expected to be called by the application every time
// when a proposal is completed or aborted by the application.
func (cs *Session) ProposalCompleted() {
	cs.assertRegularSession()
	if cs.SeriesID == cs.RespondedTo+1 {
		cs.RespondedTo = cs.SeriesID
		cs.SeriesID++
	} else {
		panic("invalid responded to/series id values")
	}
}

func (cs *Session) assertRegularSession() {
	if cs.ClientID == NotSessionManagedClientID ||
		cs.SeriesID == NoOPSeriesID {
		panic("not a regular session")
	}
}

// ValidForProposal checks whether the client session object is valid for
// making proposals.
func (cs *Session) ValidForProposal(clusterID uint64) bool {
	if cs.SeriesID == NoOPSeriesID && cs.ClientID == NotSessionManagedClientID {
		return false
	}
	if cs.ClusterID != clusterID {
		return false
	}
	if cs.ClientID == NotSessionManagedClientID {
		return false
	}
	if cs.SeriesID == SeriesIDForRegister ||
		cs.SeriesID == SeriesIDForUnregister {
		return false
	}
	if cs.RespondedTo > cs.SeriesID {
		panic("cs.RespondedTo > cs.SeriesID")
	}
	return true
}

// ValidForSessionOp checks whether the client session is valid for
// making client session related proposals, e.g. registering or unregistering
// a client session.
func (cs *Session) ValidForSessionOp(clusterID uint64) bool {
	if cs.ClusterID != clusterID {
		return false
	}
	if cs.ClientID == NotSessionManagedClientID ||
		cs.SeriesID == NoOPSeriesID {
		return false
	}
	if cs.SeriesID == SeriesIDForRegister ||
		cs.SeriesID == SeriesIDForUnregister {
		return true
	}
	return false
}
