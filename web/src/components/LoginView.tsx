import { LoginForm, UserForm } from "./AuthForms";

// LoginView renders local login and optional registration.
export function LoginView({
  allowRegistration,
  loginError,
  onLogin,
  onCreateUser,
  onClearError,
}: {
  allowRegistration: boolean;
  loginError: string | null;
  onLogin: (email: string, password: string) => void | Promise<void>;
  onCreateUser: (email: string, password: string) => void | Promise<void>;
  onClearError: () => void;
}) {
  return (
    <section className="grid two">
      <div className="panel">
        <h3>Log in</h3>
        {/* Local login form */}
        <LoginForm
          onLogin={onLogin}
          error={loginError}
          onClearError={onClearError}
        />
      </div>
      {allowRegistration ? (
        <div className="panel">
          <h3>Create user</h3>
          <UserForm onCreate={onCreateUser} />
        </div>
      ) : (
        <div className="panel">
          <h3>Registration disabled</h3>
          <p className="muted small">
            Ask an admin to create an account for you.
          </p>
        </div>
      )}
    </section>
  );
}
