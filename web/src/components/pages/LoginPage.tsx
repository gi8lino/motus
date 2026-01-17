import { LoginForm, UserForm } from "./../auth/AuthForm";
import { UI_TEXT } from "../../utils/uiText";

export type LoginViewData = {
  allowRegistration: boolean;
  loginError: string | null;
};

export type LoginViewActions = {
  onLogin: (email: string, password: string) => void | Promise<void>;
  onCreateUser: (email: string, password: string) => void | Promise<void>;
  onClearError: () => void;
};

// LoginView renders local login and optional registration.
export function LoginView({
  data,
  actions,
}: {
  data: LoginViewData;
  actions: LoginViewActions;
}) {
  const { allowRegistration, loginError } = data;
  const { onLogin, onCreateUser, onClearError } = actions;
  return (
    <section className="grid two">
      <div className="panel">
        <h3>{UI_TEXT.pages.auth.loginTitle}</h3>
        {/* Local login form */}
        <LoginForm
          onLogin={onLogin}
          error={loginError}
          onClearError={onClearError}
        />
      </div>
      {allowRegistration ? (
        <div className="panel">
          <h3>{UI_TEXT.pages.auth.createUserTitle}</h3>
          <UserForm onCreate={onCreateUser} />
        </div>
      ) : (
        <div className="panel">
          <h3>{UI_TEXT.pages.auth.registrationDisabledTitle}</h3>
          <p className="muted small hint">
            {UI_TEXT.pages.auth.registrationDisabledHint}
          </p>
        </div>
      )}
    </section>
  );
}
