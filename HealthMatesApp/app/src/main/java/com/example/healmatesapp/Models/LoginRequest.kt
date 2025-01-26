package com.example.healmatesapp.Models

data class LoginRequest(
    val login: String,
    val hashPassword: String,
    val isRemember: Boolean
)