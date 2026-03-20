# reads lines from STDIN, each line must contain an email (or alias)
# followed by a password separated by whitespace.  For each entry the
# script inserts a user record using the application's normal
# changeset so the password is hashed by Pow exactly the same way as the
# rest of the system.
#
# Example usage from the project root:
#
#   cat users.txt | \
#     docker compose exec -T backend mix run /app/priv/batch_create_users.exs
#
# The `-T` flag disables pseudo‑tty allocation which is required when
# piping input into a container.

import Ecto.Query
alias DragnCards.Users.User
alias DragnCards.Repo

IO.puts("batch_create_users running, reading from stdin...")

for raw <- IO.stream(:stdio, :line) do
  line = String.trim(raw)
  if line == "" do
    :ok
  else
    # split on whitespace: email password [alias]
    # alias is optional and defaults to the email address when omitted
    parts = String.split(line, ~r/\s+/, parts: 3)
    email      = Enum.at(parts, 0)
    password   = Enum.at(parts, 1) || ""
    user_alias = case Enum.at(parts, 2) do
      nil   -> email
      ""    -> email
      value -> value
    end

    attrs = %{
      alias: user_alias,
      email: email,
      password: password,
      password_confirmation: password,
      supporter_level: 0,
      language: "English",
      plugin_settings: %{}
    }

    changeset = User.changeset(%User{}, attrs)

    case Repo.insert(changeset) do
      {:ok, user} ->
        IO.puts("created user #{user.email}")

        confirm_time = DateTime.utc_now()

        from(p in User,
          where: p.id == ^user.id,
          update: [set: [email_confirmed_at: ^confirm_time]]
        )
        |> Repo.update_all([])
        |> case do
          {1, nil} ->
            IO.puts("email confirmed for #{user.email}")

          _ ->
            IO.puts("email NOT confirmed for #{user.email}")
        end

      {:error, changeset} ->
        IO.puts("failed to create #{email}: ")
        IO.inspect(changeset.errors)
    end
  end
end
