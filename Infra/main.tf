# Configure the Azure Provider
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.0"
    }
  }
}

# Configure the Microsoft Azure Provider
provider "azurerm" {
  features {}
  skip_provider_registration = true
}

# Create a resource group
resource "azurerm_resource_group" "aks_rg" {
  name     = "rg-aks-test"
  location = "East US"

  tags = {
    Environment = "Testing"
  }
}

# Create Azure Kubernetes Service - Simple Testing Configuration
resource "azurerm_kubernetes_cluster" "aks" {
  name                = "aks-test-cluster"
  location            = azurerm_resource_group.aks_rg.location
  resource_group_name = azurerm_resource_group.aks_rg.name
  dns_prefix          = "akstest"

  # Default node pool - minimal configuration for testing
  default_node_pool {
    name       = "default"
    node_count = 1
    vm_size    = "Standard_B2s"  # Cheaper VM size for testing
  }

  # System-assigned managed identity
  identity {
    type = "SystemAssigned"
  }

  tags = {
    Environment = "Testing"
  }
}

# Output values
output "kube_config" {
  value     = azurerm_kubernetes_cluster.aks.kube_config_raw
  sensitive = true
}

output "cluster_name" {
  value = azurerm_kubernetes_cluster.aks.name
}

output "resource_group_name" {
  value = azurerm_resource_group.aks_rg.name
}