package com.example.healmatesapp.API.AuthApiService

import com.example.healmatesapp.Models.LoginRequest
import com.example.healmatesapp.Models.AuthResponse
import retrofit2.Call
import retrofit2.http.Body
import retrofit2.http.POST

interface AuthApi {
    @POST("/register")
    fun register(@Body request: LoginRequest): Call<AuthResponse>

    @POST("/login")
    fun login(@Body request: LoginRequest): Call<AuthResponse>
}