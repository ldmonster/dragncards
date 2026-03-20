defmodule DragnCardsWeb.PowMailer do
  @moduledoc """
  Interface for Pow to send Emails with.  Uses Swoosh.
  """
  use Pow.Phoenix.Mailer
  use Swoosh.Mailer, otp_app: :dragncards

  import Swoosh.Email

  require Logger

  def cast(%{user: user, subject: subject, text: text, html: html}) do
    %Swoosh.Email{}
    |> to({"", user.email})
    |> from({"DragnCards", "postmaster@noreply.dragncards.com"})
    |> subject(subject)
    |> html_body(html)
    |> text_body(text)
  end

  def process(email) do
    case Application.get_env(:dragncards, __MODULE__, [])[:adapter] do
      nil ->
        # No mailer adapter configured (e.g. offline mode).
        # Log the email instead of trying to deliver it so registration and
        # password-reset requests still succeed without a real mail server.
        Logger.info("[Mailer] No adapter configured – skipping delivery. " <>
          "Recipient: #{inspect(email.to)}, Subject: #{inspect(email.subject)}")
        {:ok, email}

      _adapter ->
        email
        |> deliver()
        |> log_warnings()
    end
  end

  defp log_warnings({:error, reason}) do
    Logger.warning("Mailer backend failed with: #{inspect(reason)}")
  end

  defp log_warnings({:ok, response}), do: {:ok, response}
end
