package com.example.spring.modules.hello_world.v1.controllers;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
@RequestMapping("/api/v1")
public class HelloControllerV1 {
    @GetMapping("/hello")
    public Map<String, Object> sayHello() {
        return Map.of(
                "status", Map.of(
                        "code", 200,
                        "message", "Hello, Spring V1 API!"
                )
        );
    }
}
