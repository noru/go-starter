package main

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
  r := gin.Default()
  r.GET("/ping", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  })
  
  r.Any("/dummy_webhook", func(c *gin.Context) {
    log.Printf("Query params: %v", c.Request.URL.Query())
    
    log.Printf("Headers: %v", c.Request.Header)
    
    body, _ := io.ReadAll(c.Request.Body)
    log.Printf("Body: %s", string(body))
    c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
    
    c.Status(http.StatusOK)
  })
  
  r.Any("/proxy/*path", func(c *gin.Context) {
    target := c.GetHeader("x-proxy-target")
    if target == "" {
      c.Status(http.StatusBadRequest)
      return
    }

    // Get path after "/proxy" and append to target
    path := c.Param("path")
    fullURL := target + path + "?" + c.Request.URL.RawQuery
    log.Printf("Proxy to URL: %s", fullURL)
    // Read body
    body, _ := io.ReadAll(c.Request.Body)
    c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

    // Create new request
    req, err := http.NewRequest(
      c.Request.Method,
      fullURL,
      bytes.NewBuffer(body),
    )
    if err != nil {
      c.Status(http.StatusInternalServerError)
      return
    }

    // Copy headers
    for k, v := range c.Request.Header {
      if k != "x-proxy-target" {
        req.Header[k] = v
      }
    }

    // Send request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
      c.Status(http.StatusBadGateway)
      return
    }
    defer resp.Body.Close()

    // Copy response headers
    for k, v := range resp.Header {
      c.Writer.Header()[k] = v
    }

    // Copy response body and status
    respBody, _ := io.ReadAll(resp.Body)
    c.Data(resp.StatusCode, resp.Header.Get("content-type"), respBody)
  })
  
  r.Run(":8080") // listen on port 8080
} 