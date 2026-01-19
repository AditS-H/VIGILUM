// ProofVerificationClient.ts - TypeScript client for proof verification API

import { Client } from './client';

/**
 * Request/Response types for proof verification API
 */

export interface GenerateChallengeRequest {
  user_id: string;
  verifier_address: string;
}

export interface GenerateChallengeResponse {
  challenge_id: string;
  issued_at: string;
  expires_at: string;
  ttl_seconds: number;
}

export interface SubmitProofRequest {
  challenge_id: string;
  proof_data: string; // hex-encoded
  timing_variance: number;
  gas_variance: number;
  proof_nonce: string;
}

export interface SubmitProofResponse {
  is_valid: boolean;
  verification_score: number;
  verification_result: string;
  risk_score_reduction: number;
  proof_id?: string;
  verified_at?: string;
  message: string;
}

export interface GetUserProofsResponse {
  user_id: string;
  proof_count: number;
  proofs: ProofInfo[];
  average_score: number;
  page_info: PaginationInfo;
}

export interface ProofInfo {
  id: string;
  proof_hash: string;
  verification_score: number;
  verified_at?: string;
  expires_at: string;
  created_at: string;
  verifier_address: string;
}

export interface GetVerificationScoreResponse {
  user_id: string;
  verification_score: number;
  proof_count: number;
  verified_proof_count: number;
  is_verified: boolean;
  last_verified_at?: string;
  risk_score: number;
}

export interface PaginationInfo {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

export interface ErrorResponse {
  error: string;
  message: string;
  status_code: number;
  timestamp: string;
}

/**
 * ProofVerificationClient provides methods to interact with proof verification API
 */
export class ProofVerificationClient {
  private client: Client;
  private baseUrl: string = '/api/v1';

  constructor(client: Client) {
    this.client = client;
  }

  /**
   * Generate a proof challenge for a user
   */
  async generateChallenge(
    userId: string,
    verifierAddress: string
  ): Promise<GenerateChallengeResponse> {
    const request: GenerateChallengeRequest = {
      user_id: userId,
      verifier_address: verifierAddress,
    };

    const response = await this.client.post<GenerateChallengeResponse>(
      `${this.baseUrl}/proofs/challenges`,
      request
    );

    return response;
  }

  /**
   * Submit a proof for verification
   */
  async submitProof(
    challengeId: string,
    proofData: Uint8Array,
    timingVariance: number,
    gasVariance: number,
    proofNonce: string
  ): Promise<SubmitProofResponse> {
    const request: SubmitProofRequest = {
      challenge_id: challengeId,
      proof_data: this.bytesToHex(proofData),
      timing_variance: timingVariance,
      gas_variance: gasVariance,
      proof_nonce: proofNonce,
    };

    const response = await this.client.post<SubmitProofResponse>(
      `${this.baseUrl}/proofs/verify`,
      request
    );

    return response;
  }

  /**
   * Get user's proofs with pagination
   */
  async getUserProofs(
    userId: string,
    page: number = 1,
    limit: number = 10
  ): Promise<GetUserProofsResponse> {
    const params = new URLSearchParams({
      user_id: userId,
      page: page.toString(),
      limit: limit.toString(),
    });

    const response = await this.client.get<GetUserProofsResponse>(
      `${this.baseUrl}/proofs?${params.toString()}`
    );

    return response;
  }

  /**
   * Get user's verification score
   */
  async getVerificationScore(userId: string): Promise<GetVerificationScoreResponse> {
    const params = new URLSearchParams({
      user_id: userId,
    });

    const response = await this.client.get<GetVerificationScoreResponse>(
      `${this.baseUrl}/verification-score?${params.toString()}`
    );

    return response;
  }

  /**
   * Get challenge status
   */
  async getChallengeStatus(challengeId: string): Promise<any> {
    const response = await this.client.get(
      `${this.baseUrl}/proofs/challenges/${challengeId}`
    );

    return response;
  }

  /**
   * Check service health
   */
  async getHealth(): Promise<any> {
    const response = await this.client.get(
      `${this.baseUrl}/health`
    );

    return response;
  }

  /**
   * Verify firewall proof (alias endpoint)
   */
  async verifyFirewallProof(request: SubmitProofRequest): Promise<SubmitProofResponse> {
    const response = await this.client.post<SubmitProofResponse>(
      `${this.baseUrl}/firewall/verify-proof`,
      request
    );

    return response;
  }

  /**
   * Helper: Convert bytes to hex string
   */
  private bytesToHex(bytes: Uint8Array): string {
    return Array.from(bytes)
      .map(byte => byte.toString(16).padStart(2, '0'))
      .join('');
  }

  /**
   * Helper: Convert hex string to bytes
   */
  static hexToBytes(hex: string): Uint8Array {
    const result = new Uint8Array(hex.length / 2);
    for (let i = 0; i < hex.length; i += 2) {
      result[i / 2] = parseInt(hex.substr(i, 2), 16);
    }
    return result;
  }
}
