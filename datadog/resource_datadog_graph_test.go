package datadog

import (
	"fmt"
	"testing"
	"strconv"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zorkian/go-datadog-api"
)

func TestAccDatadogGraph_Basic(t *testing.T) {
	var resp datadog.Graph

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogGraphDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckDatadogGraphConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGraphExists("datadog_graph.bar", &resp),
					// TODO: Test request attributes
					resource.TestCheckResourceAttr(
						"datadog_dashboard.foo", "title", "title for dashboard foo"),
					resource.TestCheckResourceAttr(
						"datadog_dashboard.foo", "description", "description for dashboard foo"),
					resource.TestCheckResourceAttr(
						"datadog_graph.bar", "title", "title for graph bar"),
					resource.TestCheckResourceAttr(
						"datadog_graph.bar", "description", "description for graph bar"),
					resource.TestCheckResourceAttr(
						"datadog_graph.bar", "viz", "timeseries"),
				),
			},
		},
	})
}

func testAccCheckDatadogGraphDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "datadog_graph" {
			continue
		}

		dashboard_id, err := getGraphDashboard(s, rs)

		if err != nil {
			return err
		}

		// See if the graph with our title is still in the dashboard
		_, err = getGraphFromDashboard(dashboard_id, rs.Primary.Attributes["title"])

		if err != nil {
			return err
		}

		return fmt.Errorf("Graph still exists")
	}

	return nil
}

func getGraphDashboard(s *terraform.State, rs *terraform.ResourceState) (string, error) {

	for _, d := range rs.Dependencies {

		rs, ok := s.RootModule().Resources[d]

		if !ok {
			return "", fmt.Errorf("Not found: %s", d)
		}

		if rs.Primary.ID == "" {
			return "", fmt.Errorf("No ID is set")
		}

		return rs.Primary.ID, nil
	}

	return "", fmt.Errorf("Failed to find dashboard in state.") // TODO: make this a little nicer?

}

func getGraphFromDashboard(id, title string) (datadog.Graph, error) {
	client := testAccProvider.Meta().(*datadog.Client)

	graph := datadog.Graph{}

	IdInt, int_err := strconv.Atoi(id)
	if int_err == nil {
		return graph, int_err
	}

	dashboard, err := client.GetDashboard(IdInt)

	if err != nil {
		return graph, fmt.Errorf("Error retrieving associated dashboard: %s", err)
	}

	for _, g := range dashboard.Graphs {
		if g.Title != title {
			continue
		}

		return g, nil
	}

	return graph, nil
}

func testAccCheckDatadogGraphExists(n string, GraphResp *datadog.Graph) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		dashboard_id, err := getGraphDashboard(s, rs)

		if err != nil {
			return err
		}

		// See if out graph is in the dashboard
		_, err = getGraphFromDashboard(dashboard_id, rs.Primary.Attributes["title"])

		if err != nil {
			return err
		}

	return nil
	}
}

const testAccCheckDatadogGraphConfig_basic = `
resource "datadog_dashboard" "foo" {
	description = "description for dashboard foo"
	title = "title for dashboard foo"
}

resource "datadog_graph" "bar" {
	title = "title for graph bar"
    dashboard_id = "${datadog_dashboard.foo.id}"
    description = "description for graph bar"
    title = "bar"
    viz =  "timeseries"
    request {
        query =  "avg:system.cpu.system{*}"
        stacked = false
    }
    request {
        query =  "avg:system.cpu.user{*}"
        stacked = false
    }
    request {
        query =  "avg:system.mem.user{*}"
        stacked = false
    }

}
`
