package com.example.healmatesapp.Models

data class User(
    val id: Int,
    val login: String,
    val hashPassword: String,
    val createdDate: String,
    val roleId: Int,
    val isRemember: Boolean
)