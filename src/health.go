package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (sp *SleepProxy) checkContainersReady(ctx context.Context) bool {
	containers, err := sp.getProjectContainers(ctx)
	if err != nil {
		log.Printf("Failed to check containers: %v", err)
		return false
	}

	if len(containers) == 0 {
		return false
	}

	for _, c := range containers {
		if c.State != "running" {
			return false
		}

		// Check health status if available
		containerJSON, err := sp.dockerClient.ContainerInspect(ctx, c.ID)
		if err != nil {
			log.Printf("Failed to inspect container %s: %v", c.Names[0], err)
			return false
		}

		// If container has health check, wait for it to be healthy
		if containerJSON.State.Health != nil {
			if containerJSON.State.Health.Status != "healthy" {
				return false
			}
		}
	}

	// Additional HTTP check for the target service
	targetURL := fmt.Sprintf("http://%s:%s/", sp.config.TargetService, sp.config.TargetPort)
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK || resp.StatusCode < 500
}
