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

package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/apcera/termtables"
	"github.com/spf13/cobra"
)

// ssosessionsCmd represents the ssosessions command
var ssosessionsCmd = &cobra.Command{
	Use:   "ssosessions [URL of CAS server (with trailing slash)]",
	Short: "Display a report about active SSO sessions in a running CAS server.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			erAndExit("[ssosessioins] needs a URL of CAS server")
		}
		casServerBaseUrl = args[0]
		displaySsoSessions()
	},
}

func init() {
	RootCmd.AddCommand(ssosessionsCmd)
}

type SsoSessionReport struct {
	TotalPrincipals   int `json:"totalPrincipals"`
	ActiveSsoSessions []struct {
		AuthenticatedPrincipal string                          `json:"authenticated_principal"`
		AuthenticationDate     string                          `json:"authentication_date_formatted"`
		NumberOfUses           int                             `json:"number_of_uses"`
		AuthenticatedServices  map[string]AuthenticatedService `json:"authenticated_services"`
	} `json:"activeSsoSessions"`
}

type AuthenticatedService struct {
	OriginalUrl string
}

func (s *casReportingService) listActiveSessions() (*SsoSessionReport, *http.Response, error) {
	activeSessions := &SsoSessionReport{}
	resp, err := s.sling.New().Path("status/ssosessions/getSsoSessions").ReceiveSuccess(activeSessions)
	return activeSessions, resp, err
}

func displaySsoSessions() {
	casReportingService := newCasReportingService(nil)
	ssoSessionReport, resp, _ := casReportingService.listActiveSessions()
	checkResponseAndExitIfNecessary(resp)

	table := termtables.CreateTable()
	table.AddTitle("CAS server " + casServerBaseUrl + " active SSO sessions")
	table.AddHeaders("User", "Authentication Date", "Number of uses", "Services")
	if ssoSessionReport.TotalPrincipals == 0 {
		infoAndExit("No active SSO sessions")
	}
	for _, s := range ssoSessionReport.ActiveSsoSessions {
		var services string
		for _, service := range s.AuthenticatedServices {
			services += service.OriginalUrl + ","
		}
		table.AddRow(s.AuthenticatedPrincipal, s.AuthenticationDate, s.NumberOfUses, strings.TrimSuffix(services, ","))
	}

	fmt.Println()
	greenPrintln(table.Render())
}
