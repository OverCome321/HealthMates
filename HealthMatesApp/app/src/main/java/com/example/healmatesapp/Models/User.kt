package com.example.healmatesapp.Models

data class User(
    val id: Int,
    val email: String,
    val phone: String,
    val createdDate: String,
    val roleId: Int,
    val isActive: Boolean,
    val lastLogin: String
)