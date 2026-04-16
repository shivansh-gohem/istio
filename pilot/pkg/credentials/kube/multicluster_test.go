// Copyright Istio Authors
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

package kube

import (
    "testing"

    "istio.io/istio/pkg/cluster"
    "istio.io/istio/pkg/kube"
    "istio.io/istio/pkg/test"
)

func TestForClusterWithRemoteCredentialsDisabled(t *testing.T) {
    // Disable remote credentials controller using Istio's test framework
    test.SetForTest(t, &enableRemoteCredentialsController, false)

    stop := test.NewStop(t)
    localClient := kube.NewFakeClient()
    localClient.RunAndWait(stop)
    remoteClient := kube.NewFakeClient()
    remoteClient.RunAndWait(stop)
   
    mc := NewFakeController()
    sc := NewMulticluster("local", mc)
    mc.Add("local", localClient, stop)
    mc.Add("remote", remoteClient, stop)

    cases := []struct {
        cluster   cluster.ID
        expectErr bool
    }{
        // Config cluster should still work — it always gets a controller
        {"local", false},
        // Remote cluster has no controller and no auth → should error
        {"remote", true},
        // Unknown cluster should still error
        {"invalid", true},
    }
   
    for _, tt := range cases {
        t.Run(string(tt.cluster), func(t *testing.T) {
            _, err := sc.ForCluster(tt.cluster)
            if (err != nil) != tt.expectErr {
                t.Fatalf("expected err=%v, got err=%v", tt.expectErr, err)
            }
        })
    }
}

func TestAuthorizeWithRemoteCredentialsDisabled(t *testing.T) {
    test.SetForTest(t, &enableRemoteCredentialsController, false)

    stop := test.NewStop(t)
    localClient := kube.NewFakeClient()
    localClient.RunAndWait(stop)
    allowIdentities(localClient, "system:serviceaccount:ns-local:sa-allowed")
   
    mc := NewFakeController()
    sc := NewMulticluster("local", mc)
    mc.Add("local", localClient, stop)

    // Config cluster should still allow authorization
    con, err := sc.ForCluster("local")
    if err != nil {
        t.Fatal(err)
    }
    if err := con.Authorize("sa-allowed", "ns-local"); err != nil {
        t.Fatalf("expected allowed, got err=%v", err)
    }
}
