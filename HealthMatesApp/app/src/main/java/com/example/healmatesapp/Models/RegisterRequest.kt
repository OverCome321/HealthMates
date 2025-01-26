package com.example.healmatesapp.Models

data class RegisterRequest(
    val login: String,
    val hashPassword: String,
    val isRemember: Boolean
)