package ces_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-provider-google/google/acctest"
)


func TestAccCESAgent_cesAgentBasicExample_update(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"random_suffix": acctest.RandString(t, 10),
	}

  	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckCESAgentDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccCESAgent_cesAgentBasicExample_full(context),
			},
			{
				ResourceName:            "google_ces_agent.ces_agent_basic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"app_id", "agent_id"},
			},
			{
				Config: testAccCESAgent_cesAgentBasicExample_update(context),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("google_ces_agent.ces_agent_basic", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:            "google_ces_agent.ces_agent_basic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"app_id", "agent_id"},
			},
		},
	})
}

func testAccCESAgent_cesAgentBasicExample_full(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_ces_app" "ces_app_for_agent" {
  app_id = "tf-test-app-id%{random_suffix}"
  location = "us"
  description = "App used as parent for CES Agent example"
  display_name = "tf-test-my-app%{random_suffix}"

  language_settings {
    default_language_code    = "en-US"
    supported_language_codes = ["es-ES", "fr-FR"]
    enable_multilingual_support = true
    fallback_action          = "escalate"
  }

  time_zone_settings {
    time_zone = "America/Los_Angeles"
  }
}

resource "google_ces_tool" "ces_tool_for_agent" {
    location       = "us"
    app            = google_ces_app.ces_app_for_agent.app_id
    tool_id        = "tool-1"
    execution_type = "SYNCHRONOUS"
    python_function {
        name = "example_function"
        python_code = "def example_function() -> int: return 0"
    }
}

resource "google_ces_toolset" "ces_toolset_for_agent" {
  toolset_id = "toolset-1"
  location = "us"
  app      = google_ces_app.ces_app_for_agent.app_id
  display_name = "Basic toolset display name"

  open_api_toolset {
    open_api_schema = <<-EOT
      openapi: 3.0.0
      info:
        title: My Sample API
        version: 1.0.0
        description: A simple API example
      servers:
        - url: https://api.example.com/v1
      paths: {}
    EOT
    ignore_unknown_fields = false
    tls_config {
        ca_certs {
          display_name="example"
          cert="ZXhhbXBsZQ=="
        }
    }
    service_directory_config {
      service = "projects/example/locations/us/namespaces/namespace/services/service"
    }
    api_authentication {
        service_agent_id_token_auth_config {
        }
    }
  }
}


resource "google_ces_agent" "ces_child_agent" {
  agent_id = "child-agent"
  location = "us"
  app      = google_ces_app.ces_app_for_agent.app_id
  display_name = "child agent"

  instruction = "You are a helpful assistant for this example."

  model_settings {
    model       = "gemini-2.5-flash"
    temperature = 0.5
  }

  llm_agent {
  }
}


resource "google_ces_guardrail" "ces_guardrail_for_agent" {
  guardrail_id = "guardrail-id"
  location     = google_ces_app.ces_app_for_agent.location
  app          = google_ces_app.ces_app_for_agent.app_id
  display_name = "Example guardrail"
  description  = "Guardrail description"
  action {
    respond_immediately  {
        responses {
            text = "Text"
            disabled = false
        }
    }
  }
  enabled = true
  model_safety  {
    safety_settings {
        category = "HARM_CATEGORY_HATE_SPEECH"
        threshold = "BLOCK_NONE"
    }
  }
}




resource "google_ces_agent" "ces_agent_basic" {
  agent_id = "tf-test-agent-id%{random_suffix}"
  location = "us"
  app      = google_ces_app.ces_app_for_agent.app_id
  display_name = "tf-test-my-agent%{random_suffix}"

  instruction = "You are a helpful assistant for this example."

  after_agent_callbacks {
    description = "Example callback"
    disabled    = true
    python_code = "def callback(context):\n    return {'override': False}"
  }

  before_agent_callbacks {
    description = "Example callback"
    disabled    = false
    python_code = "def callback(context):\n    return {'override': False}"
  }

  after_model_callbacks {
    description = "Example callback"
    disabled    = true
    python_code = "def callback(context):\n    return {'override': False}"
  }

  before_model_callbacks {
    description = "Example callback"
    disabled    = true
    python_code = "def callback(context):\n    return {'override': False}"
  }

  after_tool_callbacks {
    description = "Example callback"
    disabled    = true
    python_code = "def callback(context):\n    return {'override': False}"
  }

  before_tool_callbacks {
    description = "Example callback"
    disabled    = true
    python_code = "def callback(context):\n    return {'override': False}"
  }

  tools = [
    google_ces_tool.ces_tool_for_agent.id
  ]

  guardrails = [
    google_ces_guardrail.ces_guardrail_for_agent.id
  ]

  toolsets {
    toolset = google_ces_toolset.ces_toolset_for_agent.id
  }

  child_agents = ["projects/${google_ces_app.ces_app_for_agent.project}/locations/us/apps/${google_ces_app.ces_app_for_agent.app_id}/agents/${google_ces_agent.ces_child_agent.agent_id}"]

  llm_agent {}
}
`, context)
}

func testAccCESAgent_cesAgentBasicExample_update(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_ces_app" "ces_app_for_agent" {
  app_id = "tf-test-app-id%{random_suffix}"
  location = "us"
  description = "App used as parent for CES Agent example"
  display_name = "tf-test-my-app%{random_suffix}"

  language_settings {
    default_language_code    = "en-US"
    supported_language_codes = ["es-ES", "fr-FR"]
    enable_multilingual_support = true
    fallback_action          = "escalate"
  }

  time_zone_settings {
    time_zone = "America/Los_Angeles"
  }
}

resource "google_ces_tool" "ces_tool_for_agent" {
    location       = "us"
    app            = google_ces_app.ces_app_for_agent.app_id
    tool_id        = "tool-1"
    execution_type = "SYNCHRONOUS"
    python_function {
        name = "example_function"
        python_code = "def example_function() -> int: return 0"
    }
}

resource "google_ces_toolset" "ces_toolset_for_agent" {
  toolset_id = "toolset-1"
  location = "us"
  app      = google_ces_app.ces_app_for_agent.app_id
  display_name = "Basic toolset display name"

  open_api_toolset {
    open_api_schema = <<-EOT
      openapi: 3.0.0
      info:
        title: My Sample API
        version: 1.0.0
        description: A simple API example
      servers:
        - url: https://api.example.com/v1
      paths: {}
    EOT
    ignore_unknown_fields = false
    tls_config {
        ca_certs {
          display_name="example"
          cert="ZXhhbXBsZQ=="
        }
    }
    service_directory_config {
      service = "projects/example/locations/us/namespaces/namespace/services/service"
    }
    api_authentication {
        service_agent_id_token_auth_config {
        }
    }
  }
}


resource "google_ces_agent" "ces_child_agent" {
  agent_id = "child-agent"
  location = "us"
  app      = google_ces_app.ces_app_for_agent.app_id
  display_name = "child agent"

  instruction = "You are a helpful assistant for this example."

  llm_agent {
  }
}


resource "google_ces_guardrail" "ces_guardrail_for_agent" {
  guardrail_id = "guardrail-id"
  location     = google_ces_app.ces_app_for_agent.location
  app          = google_ces_app.ces_app_for_agent.app_id
  display_name = "Example guardrail"
  description  = "Guardrail description"
  action {
    respond_immediately  {
        responses {
            text = "Text"
            disabled = false
        }
    }
  }
  enabled = true
  model_safety  {
    safety_settings {
        category = "HARM_CATEGORY_HATE_SPEECH"
        threshold = "BLOCK_NONE"
    }
  }
}




resource "google_ces_agent" "ces_agent_basic" {
  agent_id = "tf-test-agent-id%{random_suffix}"
  location = "us"
  app      = google_ces_app.ces_app_for_agent.app_id
  display_name = "tf-test-my-agent%{random_suffix}"

  instruction = "You are a helpful assistant for this example updated."

  after_agent_callbacks {
    description = "Example callback updated"
    disabled    = false
    python_code = "def callback(context):\n    return {'override': False}"
  }

  before_agent_callbacks {
    description = "Example callback updated"
    disabled    = true
    python_code = "def callback(context):\n    return {'override': False}"
  }

  after_model_callbacks {
    description = "Example callback updated"
    disabled    = false
    python_code = "def callback(context):\n    return {'override': False}"
  }

  before_model_callbacks {
    description = "Example callback updated"
    disabled    = false
    python_code = "def callback(context):\n    return {'override': False}"
  }

  after_tool_callbacks {
    description = "Example callback updated"
    disabled    = false
    python_code = "def callback(context):\n    return {'override': False}"
  }

  before_tool_callbacks {
    description = "Example callback updated"
    disabled    = false
    python_code = "def callback(context):\n    return {'override': False}"
  }

  tools = [
    google_ces_tool.ces_tool_for_agent.id
  ]

  guardrails = [
    google_ces_guardrail.ces_guardrail_for_agent.id
  ]

  toolsets {
    toolset = google_ces_toolset.ces_toolset_for_agent.id
  }

  child_agents = ["projects/${google_ces_app.ces_app_for_agent.project}/locations/us/apps/${google_ces_app.ces_app_for_agent.app_id}/agents/${google_ces_agent.ces_child_agent.agent_id}"]

  llm_agent {}
}
`, context)
}

func TestAccCESAgent_cesAgentRemoteDialogflowAgentExample_update(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"random_suffix": acctest.RandString(t, 10),
	}

acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckCESAgentDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccCESAgent_cesAgentRemoteDialogflowAgentExample_full(context),
			},
			{
				ResourceName:            "google_ces_agent.ces_agent_remote_dialogflow_agent",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"app_id", "agent_id"},
			},
			{
				Config: testAccCESAgent_cesAgentRemoteDialogflowAgentExample_update(context),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("google_ces_agent.ces_agent_remote_dialogflow_agent", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:            "google_ces_agent.ces_agent_remote_dialogflow_agent",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"app_id", "agent_id"},
			},
		},
	})
}

func testAccCESAgent_cesAgentRemoteDialogflowAgentExample_full(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_ces_app" "ces_app_for_agent" {
  app_id = "tf-test-app-id%{random_suffix}"
  location = "us"
  description = "App used as parent for CES Agent example"
  display_name = "tf-test-my-app%{random_suffix}"

  language_settings {
    default_language_code    = "en-US"
    supported_language_codes = ["es-ES", "fr-FR"]
    enable_multilingual_support = true
    fallback_action          = "escalate"
  }

  time_zone_settings {
    time_zone = "America/Los_Angeles"
  }
}

resource "google_ces_agent" "ces_agent_remote_dialogflow_agent" {
  agent_id = "tf-test-agent-id%{random_suffix}"
  location = "us"
  app      = google_ces_app.ces_app_for_agent.app_id
  display_name = "tf-test-my-agent%{random_suffix}"

  model_settings {
    model       = "gemini-1.5-flash"
    temperature = 0.5
  }

  remote_dialogflow_agent {
    agent = "projects/example/locations/us/agents/fake-agent"
    flow_id = "fake-flow"
    environment_id = "fake-env"
    input_variable_mapping = {
        "example" : 1
    }
    output_variable_mapping = {
        "example" : 1
    }
  }
}
`, context)
}

func testAccCESAgent_cesAgentRemoteDialogflowAgentExample_update(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_ces_app" "ces_app_for_agent" {
  app_id = "tf-test-app-id%{random_suffix}"
  location = "us"
  description = "App used as parent for CES Agent example"
  display_name = "tf-test-my-app%{random_suffix}"

  model_settings {
    model       = "gemini-2.5-flash"
    temperature = 0.7
  }

  language_settings {
    default_language_code    = "en-US"
    supported_language_codes = ["es-ES", "fr-FR"]
    enable_multilingual_support = true
    fallback_action          = "escalate"
  }

  time_zone_settings {
    time_zone = "America/Los_Angeles"
  }
}

resource "google_ces_agent" "ces_agent_remote_dialogflow_agent" {
  agent_id = "tf-test-agent-id%{random_suffix}"
  location = "us"
  app      = google_ces_app.ces_app_for_agent.app_id
  display_name = "tf-test-my-agent%{random_suffix}"

  model_settings {
    model       = "gemini-2.5-flash-lite"
    temperature = 0.7
  }


  remote_dialogflow_agent {
    agent = "projects/example/locations/us/agents/fake-agent-updated"
    flow_id = "fake-flow-updated"
    environment_id = "fake-env-updated"
    input_variable_mapping = {
        "example" : 2
    }
    output_variable_mapping = {
        "example" : 2
    }
  }
}
`, context)
}
