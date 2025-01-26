package com.example.healmatesapp.VM

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import com.example.healmatesapp.Models.LoginRequest
import com.example.healmatesapp.Models.AuthResponse
import com.example.healmatesapp.RetrofitClientService.RetrofitClient
import retrofit2.Call
import retrofit2.Callback
import retrofit2.Response

class LoginViewModel : ViewModel() {

    // LiveData для хранения результата авторизации
    private val _loginResult = MutableLiveData<String>()
    val loginResult: LiveData<String> get() = _loginResult

    // LiveData для хранения ошибок
    private val _errorMessage = MutableLiveData<String>()
    val errorMessage: LiveData<String> get() = _errorMessage

    // Функция для выполнения авторизации
    fun login(login: String, password: String) {
        val request = LoginRequest(login, password, isRemember = true)

        RetrofitClient.instance.login(request).enqueue(object : Callback<AuthResponse> {
            override fun onResponse(call: Call<AuthResponse>, response: Response<AuthResponse>) {
                if (response.isSuccessful) {
                    val token = response.body()?.token
                    _loginResult.value = "Login successful! Token: $token"
                } else {
                    _errorMessage.value = "Login failed"
                }
            }

            override fun onFailure(call: Call<AuthResponse>, t: Throwable) {
                _errorMessage.value = "Network error: ${t.message}"
            }
        })
    }
}