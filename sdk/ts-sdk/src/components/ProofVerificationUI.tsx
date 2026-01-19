// ProofVerificationUI.tsx - React components for proof verification interface

import React, { useEffect, useMemo, useState } from 'react';
import { ProofVerificationClient } from '../proof-client';

/**
 * ProofVerificationPage - Main page component for proof verification
 */
export const ProofVerificationPage: React.FC<{ userId: string; verifierAddress: string }> = ({
  userId,
  verifierAddress,
}) => {
  const proofClient = useMemo(() => new ProofVerificationClient({ baseUrl: '/api/v1' }), []);

  return (
    <div className="proof-verification-container">
      <h1>Proof Verification System</h1>
      <div className="grid grid-cols-2 gap-4">
        <ChallengeGeneratorCard
          userId={userId}
          verifierAddress={verifierAddress}
          proofClient={proofClient}
        />
        <VerificationScoreCard userId={userId} proofClient={proofClient} />
      </div>
      <div className="mt-8">
        <UserProofsHistory userId={userId} proofClient={proofClient} />
      </div>
    </div>
  );
};

/**
 * ChallengeGeneratorCard - Component to generate and display proof challenges
 */
const ChallengeGeneratorCard: React.FC<{
  userId: string;
  verifierAddress: string;
  proofClient: ProofVerificationClient;
}> = ({ userId, verifierAddress, proofClient }) => {
  const [challengeId, setChallengeId] = useState<string | null>(null);
  const [expiresAt, setExpiresAt] = useState<Date | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [timeRemaining, setTimeRemaining] = useState<string>('');

  // Generate challenge
  const handleGenerateChallenge = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await proofClient.generateChallenge(userId, verifierAddress);
      setChallengeId(response.challenge_id);
      setExpiresAt(new Date(response.expires_at));
    } catch (err: any) {
      setError(err.message || 'Failed to generate challenge');
    } finally {
      setLoading(false);
    }
  };

  // Update countdown timer
  useEffect(() => {
    if (!expiresAt) return;

    const interval = setInterval(() => {
      const now = new Date();
      const diff = expiresAt.getTime() - now.getTime();

      if (diff <= 0) {
        setTimeRemaining('Expired');
        clearInterval(interval);
      } else {
        const minutes = Math.floor(diff / 60000);
        const seconds = Math.floor((diff % 60000) / 1000);
        setTimeRemaining(`${minutes}m ${seconds}s`);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [expiresAt]);

  return (
    <div className="card bg-white rounded-lg shadow p-6">
      <h2 className="text-2xl font-bold mb-4">Generate Proof Challenge</h2>

      {error && (
        <div className="alert alert-error mb-4">
          <span>{error}</span>
        </div>
      )}

      <div className="form-group mb-4">
        <label className="label">
          <span className="label-text">User ID</span>
        </label>
        <input
          type="text"
          value={userId}
          disabled
          className="input input-bordered w-full"
        />
      </div>

      <div className="form-group mb-6">
        <label className="label">
          <span className="label-text">Verifier Address</span>
        </label>
        <input
          type="text"
          value={verifierAddress}
          disabled
          className="input input-bordered w-full"
        />
      </div>

      {challengeId ? (
        <div className="alert alert-success mb-4">
          <div>
            <h3 className="font-bold">Challenge Generated!</h3>
            <div className="text-sm mt-2">
              <p>
                <strong>Challenge ID:</strong> {challengeId}
              </p>
              <p>
                <strong>Expires In:</strong> {timeRemaining}
              </p>
            </div>
          </div>
        </div>
      ) : null}

      <button
        onClick={handleGenerateChallenge}
        disabled={loading || !!challengeId}
        className="btn btn-primary w-full"
      >
        {loading ? 'Generating...' : 'Generate Challenge'}
      </button>

      {challengeId && (
        <ProofSubmissionForm
          challengeId={challengeId}
          proofClient={proofClient}
          onSubmitSuccess={() => {
            setChallengeId(null);
            setExpiresAt(null);
          }}
        />
      )}
    </div>
  );
};

/**
 * ProofSubmissionForm - Component for submitting proofs
 */
const ProofSubmissionForm: React.FC<{
  challengeId: string;
  proofClient: ProofVerificationClient;
  onSubmitSuccess: () => void;
}> = ({ challengeId, proofClient, onSubmitSuccess }) => {
  const [proofData, setProofData] = useState<string>('');
  const [timingVariance, setTimingVariance] = useState<number>(150);
  const [gasVariance, setGasVariance] = useState<number>(800);
  const [proofNonce, setProofNonce] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<any>(null);

  // Generate random nonce
  useEffect(() => {
    setProofNonce(Math.random().toString(36).substring(2, 15));
  }, [challengeId]);

  const handleSubmitProof = async () => {
    setLoading(true);
    setError(null);

    try {
      if (!proofData) {
        throw new Error('Proof data is required');
      }

      const normalizedHex = proofData.trim().replace(/^0x/i, '');
      if (!normalizedHex) {
        throw new Error('Proof data is required');
      }
      if (normalizedHex.length % 2 !== 0) {
        throw new Error('Proof data must have an even number of hex characters');
      }
      if (!/^[0-9a-fA-F]+$/.test(normalizedHex)) {
        throw new Error('Proof data must be valid hex characters');
      }

      const proofBytes = ProofVerificationClient.hexToBytes(normalizedHex);
      const response = await proofClient.submitProof(
        challengeId,
        proofBytes,
        timingVariance,
        gasVariance,
        proofNonce
      );

      setResult(response);

      if (response.is_valid) {
        setTimeout(() => {
          onSubmitSuccess();
        }, 2000);
      }
    } catch (err: any) {
      setError(err.message || 'Failed to submit proof');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mt-6 p-4 border border-gray-200 rounded">
      <h3 className="text-xl font-bold mb-4">Submit Proof</h3>

      {result ? (
        <div className={`alert alert-${result.is_valid ? 'success' : 'error'} mb-4`}>
          <div>
            <h4 className="font-bold">{result.verification_result}</h4>
            <div className="text-sm mt-2">
              <p>
                <strong>Score:</strong> {(result.verification_score * 100).toFixed(1)}%
              </p>
              <p>
                <strong>Risk Reduction:</strong> {result.risk_score_reduction} points
              </p>
            </div>
          </div>
        </div>
      ) : null}

      {error && (
        <div className="alert alert-error mb-4">
          <span>{error}</span>
        </div>
      )}

      <div className="form-group mb-4">
        <label className="label">
          <span className="label-text">Proof Data (hex)</span>
        </label>
        <textarea
          value={proofData}
          onChange={(e) => setProofData(e.target.value)}
          placeholder="00010203040506070809..."
          className="textarea textarea-bordered w-full"
          rows={3}
        />
      </div>

      <div className="grid grid-cols-2 gap-4 mb-4">
        <div className="form-group">
          <label className="label">
            <span className="label-text">Timing Variance (ms)</span>
          </label>
          <input
            type="number"
            value={timingVariance}
            onChange={(e) => setTimingVariance(Number(e.target.value))}
            className="input input-bordered w-full"
          />
        </div>
        <div className="form-group">
          <label className="label">
            <span className="label-text">Gas Variance (units)</span>
          </label>
          <input
            type="number"
            value={gasVariance}
            onChange={(e) => setGasVariance(Number(e.target.value))}
            className="input input-bordered w-full"
          />
        </div>
      </div>

      <div className="form-group mb-6">
        <label className="label">
          <span className="label-text">Proof Nonce</span>
        </label>
        <input
          type="text"
          value={proofNonce}
          disabled
          className="input input-bordered w-full"
        />
      </div>

      <button
        onClick={handleSubmitProof}
        disabled={loading || !proofData}
        className="btn btn-success w-full"
      >
        {loading ? 'Verifying...' : 'Submit Proof'}
      </button>
    </div>
  );
};

/**
 * VerificationScoreCard - Component to display user's verification score
 */
const VerificationScoreCard: React.FC<{
  userId: string;
  proofClient: ProofVerificationClient;
}> = ({ userId, proofClient }) => {
  const [score, setScore] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchScore = async () => {
      setLoading(true);
      setError(null);

      try {
        const response = await proofClient.getVerificationScore(userId);
        setScore(response);
      } catch (err: any) {
        setError(err.message || 'Failed to fetch verification score');
      } finally {
        setLoading(false);
      }
    };

    fetchScore();
    const interval = setInterval(fetchScore, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  }, [userId, proofClient]);

  if (loading) {
    return <div className="card bg-white rounded-lg shadow p-6">Loading...</div>;
  }

  if (error) {
    return (
      <div className="card bg-white rounded-lg shadow p-6">
        <div className="alert alert-error">
          <span>{error}</span>
        </div>
      </div>
    );
  }

  const scorePercentage = (score.verification_score * 100).toFixed(1);
  const riskLevel = score.risk_score > 70 ? 'High' : score.risk_score > 40 ? 'Medium' : 'Low';

  return (
    <div className="card bg-white rounded-lg shadow p-6">
      <h2 className="text-2xl font-bold mb-4">Verification Status</h2>

      <div className="stat">
        <div className="stat-title">Verification Score</div>
        <div className="stat-value text-primary">{scorePercentage}%</div>
      </div>

      <div className="stat">
        <div className="stat-title">Status</div>
        <div className="stat-value text-lg">
          {score.is_verified ? (
            <span className="badge badge-success">Verified</span>
          ) : (
            <span className="badge badge-warning">Unverified</span>
          )}
        </div>
      </div>

      <div className="stat">
        <div className="stat-title">Risk Score</div>
        <div className="stat-value text-lg">{score.risk_score}</div>
        <div className="text-sm text-gray-500">{riskLevel} Risk Level</div>
      </div>

      <div className="stat">
        <div className="stat-title">Successful Proofs</div>
        <div className="stat-value">{score.verified_proof_count}</div>
        <div className="text-sm text-gray-500">of {score.proof_count} total</div>
      </div>

      {score.last_verified_at && (
        <div className="stat">
          <div className="stat-title">Last Verified</div>
          <div className="text-sm">
            {new Date(score.last_verified_at).toLocaleDateString()}
          </div>
        </div>
      )}
    </div>
  );
};

/**
 * UserProofsHistory - Component to display user's proof history
 */
const UserProofsHistory: React.FC<{
  userId: string;
  proofClient: ProofVerificationClient;
}> = ({ userId, proofClient }) => {
  const [proofs, setProofs] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);

  useEffect(() => {
    const fetchProofs = async () => {
      setLoading(true);
      setError(null);

      try {
        const response = await proofClient.getUserProofs(userId, page, 10);
        setProofs(response);
      } catch (err: any) {
        setError(err.message || 'Failed to fetch proofs');
      } finally {
        setLoading(false);
      }
    };

    fetchProofs();
  }, [userId, page, proofClient]);

  if (loading) {
    return <div className="text-center p-4">Loading proofs...</div>;
  }

  if (error) {
    return (
      <div className="alert alert-error">
        <span>{error}</span>
      </div>
    );
  }

  if (!proofs || !Array.isArray(proofs.proofs)) {
    return (
      <div className="alert alert-warning">
        <span>No proofs found.</span>
      </div>
    );
  }

  const totalPages = proofs.page_info?.total_pages ?? 1;

  return (
    <div className="card bg-white rounded-lg shadow p-6">
      <h2 className="text-2xl font-bold mb-4">Proof History</h2>

      <div className="stat">
        <div className="stat-title">Total Proofs</div>
        <div className="stat-value">{proofs.proof_count}</div>
        <div className="text-sm">Average Score: {(proofs.average_score * 100).toFixed(1)}%</div>
      </div>

      <div className="overflow-x-auto mt-4">
        <table className="table w-full">
          <thead>
            <tr>
              <th>Proof ID</th>
              <th>Status</th>
              <th>Score</th>
              <th>Verified At</th>
              <th>Expires At</th>
            </tr>
          </thead>
          <tbody>
            {proofs.proofs.map((proof: any) => (
              <tr key={proof.id}>
                <td className="text-sm font-mono">{proof.id.substring(0, 12)}...</td>
                <td>
                  {proof.verified_at ? (
                    <span className="badge badge-success">Verified</span>
                  ) : (
                    <span className="badge badge-warning">Pending</span>
                  )}
                </td>
                <td>{(proof.verification_score * 100).toFixed(1)}%</td>
                <td>
                  {proof.verified_at
                    ? new Date(proof.verified_at).toLocaleDateString()
                    : '-'}
                </td>
                <td>{new Date(proof.expires_at).toLocaleDateString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex gap-2 mt-4 justify-center">
        <button
          onClick={() => setPage(Math.max(1, page - 1))}
          disabled={page === 1}
          className="btn btn-sm"
        >
          Previous
        </button>
        <span className="flex items-center">
          Page {page} of {totalPages}
        </span>
        <button
          onClick={() => setPage(Math.min(totalPages, page + 1))}
          disabled={page === totalPages}
          className="btn btn-sm"
        >
          Next
        </button>
      </div>
    </div>
  );
};

export default ProofVerificationPage;
