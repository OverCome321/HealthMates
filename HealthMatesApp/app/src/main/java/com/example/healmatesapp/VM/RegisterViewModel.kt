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

class RegisterViewModel : ViewModel() {
    private val _registerResult = MutableLiveData<String>()
    val registerResult: LiveData<String> get() = _registerResult

    private val _errorMessage = MutableLiveData<String>()
    val errorMessage: LiveData<String> get() = _errorMessage

    fun register(email: String, password: String) {
        val request = LoginRequest(email, password, isRemember = false)

        RetrofitClient.instance.register(request).enqueue(object : Callback<AuthResponse> {
            override fun onResponse(call: Call<AuthResponse>, response: Response<AuthResponse>) {
                if (response.isSuccessful) {
                    _registerResult.postValue("Registration successful!")
                } else {
                    _errorMessage.postValue("Error: ${response.code()}")
                }
            }

            override fun onFailure(call: Call<AuthResponse>, t: Throwable) {
                _errorMessage.postValue("Failure: ${t.message}")
            }
        })
    }
}