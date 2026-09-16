def test_login_page_renders_form(page, frontend_url):
    page.goto(frontend_url)

    login_input = page.get_by_placeholder("Введите логин")
    password_input = page.get_by_placeholder("Введите пароль")
    submit_button = page.get_by_role("button", name="Войти")

    login_input.wait_for(state="visible")
    assert login_input.is_visible()
    assert password_input.is_visible()
    assert submit_button.is_visible()


def test_submit_is_disabled_until_both_fields_are_filled(page, frontend_url):
    page.goto(frontend_url)

    submit_button = page.get_by_role("button", name="Войти")
    submit_button.wait_for(state="visible")
    assert submit_button.is_disabled()

    page.get_by_placeholder("Введите логин").fill("someone")
    page.get_by_placeholder("Введите пароль").fill("wrong-password")
    assert submit_button.is_enabled()


def test_invalid_credentials_show_error_message(page, frontend_url, gateway_url):
    # Needs a live gateway/employee-service to actually answer 401 — otherwise
    # the UI shows a generic network-error state instead of this message.
    page.goto(frontend_url)

    page.get_by_placeholder("Введите логин").fill("qa-nonexistent-user")
    page.get_by_placeholder("Введите пароль").fill("definitely-wrong")
    page.get_by_role("button", name="Войти").click()

    error = page.get_by_text("Неверный логин или пароль")
    error.wait_for(state="visible", timeout=10_000)
