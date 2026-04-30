import { Route } from "react-router-dom";
import { lazy } from "react";
import Landing from "../../pages/Landing";
import Login from "../../pages/Login";
const Register = lazy(() => import("../../pages/Register"));
const ForgotPassword = lazy(() => import("../../pages/ForgotPassword"));
const ResetPassword = lazy(() => import("../../pages/ResetPassword"));
const OAuthSuccess = lazy(() => import("../../pages/OAuthSuccess"));
const OAuthError = lazy(() => import("../../pages/OAuthError"));
const NotFound = lazy(() => import("../../pages/NotFound"));

/**
 * Public routes accessible to all users (authenticated or not)
 */
export const PublicRoutes = (
  <>
    <Route path="/" element={<Landing />} />
    <Route path="/login" element={<Login />} />
    <Route path="/register" element={<Register />} />
    <Route path="/forgot-password" element={<ForgotPassword />} />
    <Route path="/reset-password" element={<ResetPassword />} />
    <Route path="/auth/oauth-success" element={<OAuthSuccess />} />
    <Route path="/auth/oauth-error" element={<OAuthError />} />
    <Route path="/404" element={<NotFound />} />
  </>
);
