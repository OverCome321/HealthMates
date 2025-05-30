package com.example.healmatesapp.API.Models

data class LoginRequest(
    val email: String,
    val phone: String,
    val password: String,
    val provider: String,
    val token: String? = null
) 