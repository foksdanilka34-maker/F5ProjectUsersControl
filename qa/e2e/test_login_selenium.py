from selenium.webdriver.common.by import By
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait


def test_login_page_title_and_form_fields(selenium_driver, frontend_url):
    selenium_driver.get(frontend_url)

    wait = WebDriverWait(selenium_driver, 10)
    wait.until(EC.presence_of_element_located((By.CSS_SELECTOR, "input[autocomplete='username']")))

    assert "F5" in selenium_driver.title or "Vite" in selenium_driver.title

    login_input = selenium_driver.find_element(By.CSS_SELECTOR, "input[autocomplete='username']")
    password_input = selenium_driver.find_element(By.CSS_SELECTOR, "input[autocomplete='current-password']")
    submit_button = selenium_driver.find_element(By.CSS_SELECTOR, "button[type='submit']")

    assert login_input.is_displayed()
    assert password_input.is_displayed()
    assert submit_button.is_displayed()
    assert submit_button.text == "Войти"


def test_typing_credentials_enables_submit_button(selenium_driver, frontend_url):
    selenium_driver.get(frontend_url)

    wait = WebDriverWait(selenium_driver, 10)
    login_input = wait.until(
        EC.presence_of_element_located((By.CSS_SELECTOR, "input[autocomplete='username']"))
    )
    password_input = selenium_driver.find_element(By.CSS_SELECTOR, "input[autocomplete='current-password']")
    submit_button = selenium_driver.find_element(By.CSS_SELECTOR, "button[type='submit']")

    assert submit_button.get_attribute("disabled") is not None

    login_input.send_keys("qa-user")
    password_input.send_keys("qa-password")

    wait.until(lambda _: submit_button.get_attribute("disabled") is None)
