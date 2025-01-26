package com.example.healmatesapp.Views.activities

import android.os.Bundle
import android.text.Editable
import android.text.TextWatcher
import android.view.MotionEvent
import android.view.inputmethod.InputMethodManager
import android.widget.Button
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.Observer
import androidx.lifecycle.ViewModelProvider
import com.example.healmatesapp.R
import com.example.healmatesapp.VM.RegisterViewModel

class RegisterActivity : AppCompatActivity() {

    private lateinit var viewModel: RegisterViewModel
    private lateinit var editTextRegEmail: EditText
    private lateinit var editTextRegPassword: EditText
    private lateinit var editTextConfirmPassword: EditText
    private lateinit var buttonRegister: Button

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_register)

        // Инициализация View-элементов
        editTextRegEmail = findViewById(R.id.editTextRegEmail)
        editTextRegPassword = findViewById(R.id.editTextRegPassword)
        editTextConfirmPassword = findViewById(R.id.editTextConfirmPassword)
        buttonRegister = findViewById(R.id.buttonRegister)

        // Инициализация ViewModel
        viewModel = ViewModelProvider(this).get(RegisterViewModel::class.java)

        // Наблюдаем за изменениями статуса регистрации
        viewModel.registerResult.observe(this, Observer { result ->
            if (result != null) {
                Toast.makeText(this, result, Toast.LENGTH_SHORT).show()
                finish()
            }
        })

        // Наблюдаем за ошибками
        viewModel.errorMessage.observe(this, Observer { error ->
            if (error != null) {
                showError(error)
            }
        })

        // Обработка нажатия по экрану для скрытия клавиатуры
        val rootLayout: LinearLayout = findViewById(R.id.rootLayout)  // Корневой LinearLayout
        rootLayout.setOnTouchListener { _, event ->
            if (event.action == MotionEvent.ACTION_DOWN) {
                // Закрыть клавиатуру при нажатии вне поля ввода
                val imm = getSystemService(INPUT_METHOD_SERVICE) as InputMethodManager
                imm.hideSoftInputFromWindow(currentFocus?.windowToken, 0)
            }
            false
        }

        // Валидация почты (только email)
        editTextRegEmail.addTextChangedListener(object : TextWatcher {
            override fun afterTextChanged(s: Editable?) {
                val email = s.toString()
                if (!android.util.Patterns.EMAIL_ADDRESS.matcher(email).matches()) {
                    editTextRegEmail.error = "Введите правильный email"
                }
            }

            override fun beforeTextChanged(s: CharSequence?, start: Int, count: Int, after: Int) {}
            override fun onTextChanged(s: CharSequence?, start: Int, before: Int, count: Int) {}
        })

        // Валидация пароля (не менее одной цифры)
        editTextRegPassword.addTextChangedListener(object : TextWatcher {
            override fun afterTextChanged(s: Editable?) {
                val password = s.toString()
                if (password.length < 6 || !password.matches(".*\\d.*".toRegex())) {
                    editTextRegPassword.error = "Пароль должен быть сложным (не менее 6 символов и содержать цифры)"
                }
            }

            override fun beforeTextChanged(s: CharSequence?, start: Int, count: Int, after: Int) {}
            override fun onTextChanged(s: CharSequence?, start: Int, before: Int, count: Int) {}
        })

        // Обработка кнопки регистрации
        buttonRegister.setOnClickListener {
            val email = editTextRegEmail.text.toString().trim()
            val password = editTextRegPassword.text.toString().trim()
            val confirmPassword = editTextConfirmPassword.text.toString().trim()

            if (validateInput(email, password, confirmPassword)) {
                registerUser(email, password)
            }
        }
    }

    private fun validateInput(
        email: String,
        password: String,
        confirmPassword: String
    ): Boolean {
        return when {
            email.isEmpty() -> {
                showError("Введите email")
                false
            }
            password.isEmpty() -> {
                showError("Введите пароль")
                false
            }
            password != confirmPassword -> {
                showError("Пароли не совпадают")
                false
            }
            else -> true
        }
    }

    private fun registerUser(email: String, password: String) {
        viewModel.register(email, password)
    }

    private fun showError(message: String) {
        Toast.makeText(this, message, Toast.LENGTH_SHORT).show()
    }
}
