package com.example.healmatesapp.API.AuthApiService

import com.example.healmatesapp.API.Models.LoginRequest
import com.example.healmatesapp.API.Models.LoginResponse
import com.example.healmatesapp.API.Models.RegisterRequest
import com.example.healmatesapp.API.Models.RegisterResponse
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.POST

interface AuthApi {
    @POST("auth/login")
    suspend fun login(@Body request: LoginRequest): Response<LoginResponse>

    @POST("auth/register")
    suspend fun register(@Body request: RegisterRequest): Response<RegisterResponse>

    @POST("auth/google")
    suspend fun loginWithGoogle(@Body request: LoginRequest): Response<LoginResponse>
} 