// Copyright 2016-2018 Authors of Cilium
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

package v3

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func getInt64(i int64) *int64 {
	return &i
}

var (
	JSONSchema = map[string]apiextensionsv1beta1.JSONSchemaProps{
		"spec":  spec,
		"specs": specs,
	}

	cidrSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "CIDR is a CIDR prefix / IP Block.",
		Type:        "string",
		OneOf: []apiextensionsv1beta1.JSONSchemaProps{
			{
				// IPv4 CIDR
				Type: "string",
				Pattern: `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4]` +
					`[0-9]|[01]?[0-9][0-9]?)\/([0-9]|[1-2][0-9]|3[0-2])$`,
			},
			//{
			//	// IPv6 CIDR
			//	Type: "string",
			//	Pattern: `^s*((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]` +
			//		`{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|` +
			//		`2[0-4]d|1dd|[1-9]?d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4})` +
			//		`{1,2})|:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3})` +
			//		`|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]` +
			//		`{1,4})?:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3}` +
			//		`))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]` +
			//		`{1,4}){0,2}:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d))` +
			//		`{3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:` +
			//		`[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|` +
			//		`1dd|[1-9]?d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|` +
			//		`((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d` +
			//		`|1dd|[1-9]?d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4})` +
			//		`{0,5}:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3}))|` +
			//		`:)))(%.+)?s*/([0-9]|[1-9][0-9]|1[0-1][0-9]|12[0-8])$`,
			//},
		},
	}

	cidrRuleSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Required: []string{
			"anyOf",
		},
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"anyOf": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &cidrSchema,
				},
			},
			"except": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &cidrSchema,
				},
			},
			"toPorts": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &portRuleSchema,
				},
			},
		},
	}

	egressRuleSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"toIdentities": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &identityRuleSchema,
				},
			},
			"toRequires": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &identityRequirementSchema,
				},
			},
			"toCIDR": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &cidrRuleSchema,
				},
			},
			"toEntities": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &entityRuleSchema,
				},
			},
			"toServices": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &serviceRuleSchema,
				},
			},
		},
	}

	entityRuleSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"anyOf": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &entitySchema,
				},
			},
			"toPorts": portRuleSchema,
		},
	}

	entitySchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Type:        "string",
		Enum: []apiextensionsv1beta1.JSON{
			{
				Raw: []byte(`"all"`),
			},
			{
				Raw: []byte(`"host"`),
			},
			{
				Raw: []byte(`"world"`),
			},
		},
	}

	identityRequirementSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"anyOf": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &identitySelectorSchema,
				},
			},
		},
	}

	identityRuleSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"identitySelector": identitySelectorSchema,
			"toPorts":          portRuleSchema,
		},
	}

	identitySelectorSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"matchLabels": {
				Description: "",
				Type:        "object",
			},
			"matchExpressions": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &labelSelectorRequirementSchema,
				},
			},
		},
	}

	ingressRuleSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"fromIdentities": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &identityRuleSchema,
				},
			},
			"fromRequires": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &identityRequirementSchema,
				},
			},
			"fromCIDR": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &cidrRuleSchema,
				},
			},
			"fromEntities": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &entityRuleSchema,
				},
			},
		},
	}

	k8sServiceNamespaceSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Required: []string{
			"serviceSelector",
		},
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"serviceSelector": {
				Type: "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &serviceSelectorSchema,
				},
			},
			"namespace": {
				Type: "string",
			},
		},
	}

	k8sServiceSelectorNamespaceSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"serviceName": {
				Type: "string",
			},
			"serviceNamespace": {
				Type: "string",
			},
		},
	}

	l7RulesSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		// FIXME confirm existence of anyOf in kube-apiserver
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"http": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &portRuleHTTPSchema,
				},
			},
			"kafka": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &portRuleKafkaSchema,
				},
			},
		},
	}

	labelSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Required: []string{
			"key",
		},
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"key": {
				Type: "string",
			},
			"source": {
				Description: "",
				Type:        "string",
			},
			"value": {
				Type: "string",
			},
		},
	}

	labelSelectorRequirementSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"key": {
				Description: "",
				Type:        "string",
			},
			"operator": {
				Description: "",
				Type:        "string",
				Enum: []apiextensionsv1beta1.JSON{
					{
						Raw: []byte(`"In"`),
					},
					{
						Raw: []byte(`"NotIn"`),
					},
					{
						Raw: []byte(`"Exists"`),
					},
					{
						Raw: []byte(`"DoesNotExist"`),
					},
				},
			},
			"values": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &apiextensionsv1beta1.JSONSchemaProps{
						Type: "string",
					},
				},
			},
		},
		Required: []string{"key", "operator"},
	}

	portProtocolSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Required: []string{
			"port",
		},
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"port": {
				Description: "Port is an L4 port number. For now the string will be strictly " +
					"parsed as a single uint16. In the future, this field may support ranges " +
					"in the form \"1024-2048",
				Type: "string",
				// uint16 string regex
				Pattern: `^(6553[0-5]|655[0-2][0-9]|65[0-4][0-9]{2}|6[0-4][0-9]{3}|` +
					`[1-5][0-9]{4}|[0-9]{1,4})$`,
			},
			"protocol": {
				Description: `Protocol is the L4 protocol. If omitted or empty, any protocol ` +
					`matches. Accepted values: "TCP", "UDP", ""/"ANY"\n\nMatching on ` +
					`ICMP is not supported.`,
				Type: "string",
				Enum: []apiextensionsv1beta1.JSON{
					{
						Raw: []byte(`"TCP"`),
					},
					{
						Raw: []byte(`"UDP"`),
					},
					{
						Raw: []byte(`"ANY"`),
					},
				},
			},
		},
	}

	portRuleHTTPSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"headers": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &apiextensionsv1beta1.JSONSchemaProps{
						Type: "string",
					},
				},
			},
			"host": {
				Description: "",
				Type:        "string",
				Format:      "idn-hostname",
			},
			"method": {
				Description: "",
				Type:        "string",
			},
			"path": {
				Description: "",
				Type:        "string",
			},
		},
	}

	portRuleKafkaSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"role": {
				Description: "",
				Type:        "string",
				Enum: []apiextensionsv1beta1.JSON{
					{
						Raw: []byte(`"produce"`),
					},
					{
						Raw: []byte(`"consume"`),
					},
				},
			},
			"apiKey": {
				Description: "",
				Type:        "string",
			},
			"apiVersion": {
				Description: "",
				Type:        "string",
			},
			"clientID": {
				Description: "",
				Type:        "string",
			},
			"topic": {
				Description: "",
				Type:        "string",
				MaxLength:   getInt64(255),
			},
		},
	}

	portRuleSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"anyOf": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &portProtocolSchema,
				},
			},
			"rules": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &l7RulesSchema,
				},
			},
		},
	}

	ruleSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Required: []string{
			"identitySelector",
		},
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"identitySelector": identitySelectorSchema,
			"Description": {
				Description: "",
				Type:        "string",
			},
			"ingress": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &ingressRuleSchema,
				},
			},
			"egress": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &egressRuleSchema,
				},
			},
			"labels": {
				Description: "",
				Type:        "array",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &labelSchema,
				},
			},
		},
	}

	serviceRuleSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"k8sServiceSelector": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &k8sServiceSelectorNamespaceSchema,
				},
			},
			"k8sService": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &k8sServiceNamespaceSchema,
				},
			},
			"toPorts": {
				Description: "",
				Type:        "object",
				Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
					Schema: &portRuleSchema,
				},
			},
		},
	}

	serviceSelectorSchema = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Type:        "object",
		Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
			Schema: &identitySelectorSchema,
		},
	}

	spec = apiextensionsv1beta1.JSONSchemaProps{
		Description: "",
		Type:        "object",
		Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
			Schema: &ruleSchema,
		},
	}

	specs = apiextensionsv1beta1.JSONSchemaProps{
		Description: "Specs is a list of desired Cilium specific rule specification.",
		Type:        "array",
		Items: &apiextensionsv1beta1.JSONSchemaPropsOrArray{
			Schema: &ruleSchema,
		},
	}
)
