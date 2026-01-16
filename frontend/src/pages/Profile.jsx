import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { FaUserCircle, FaQrcode, FaShareAlt, FaEdit, FaSignOutAlt } from 'react-icons/fa';
import api from '../services/api';

const Profile = () => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(true);
    const [profile, setProfile] = useState(null);
    const [isEditing, setIsEditing] = useState(false);
    const [shareData, setShareData] = useState(null);
    const [error, setError] = useState('');

    const [formData, setFormData] = useState({
        full_name: '',
        date_of_birth: '',
        gender: 'Other',
        phone: '',
        address: ''
    });

    useEffect(() => {
        fetchProfile();
    }, []);

    const fetchProfile = async () => {
        try {
            const response = await api.get('/patients/profile');
            const profileData = response.data.data;
            setProfile(profileData);
            setFormData({
                full_name: profileData.full_name,
                date_of_birth: profileData.date_of_birth.split('T')[0],
                gender: profileData.gender,
                phone: profileData.phone,
                address: profileData.address
            });
            if (profileData.share_token) {
                setShareData({
                    share_token: profileData.share_token,
                    share_url: `/api/patients/share/${profileData.share_token}` // Just for local reference
                });
            }
        } catch (err) {
            if (err.response?.status === 404) {
                setProfile(null);
                setIsEditing(true); // Force create mode
            } else if (err.response?.status === 401) {
                navigate('/login');
            } else {
                setError('Failed to load profile');
            }
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async (e) => {
        e.preventDefault();
        try {
            if (profile) {
                await api.put('/patients/profile', formData);
            } else {
                await api.post('/patients/profile', formData);
            }
            setIsEditing(false);
            fetchProfile();
        } catch (err) {
            setError(err.response?.data?.error || 'Failed to save profile');
        }
    };

    const generateShareLink = async () => {
        try {
            const response = await api.post('/patients/share');
            setShareData(response.data.data);
        } catch (err) {
            setError('Failed to generate share link');
        }
    };

    const handleLogout = () => {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        navigate('/login');
    };

    if (loading) return <div className="container" style={{ padding: '2rem', textAlign: 'center' }}>Loading...</div>;

    return (
        <div style={{ minHeight: '100vh', background: 'var(--bg-light)' }}>
            {/* Navbar */}
            <nav style={{ background: 'var(--white)', padding: '1rem 0', boxShadow: 'var(--shadow-sm)' }}>
                <div className="container" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <h1 style={{ color: 'var(--primary)', margin: 0, fontSize: '1.5rem', fontWeight: 700 }}>Health Bar</h1>
                    <button onClick={handleLogout} className="btn btn-outline" style={{ fontSize: '0.875rem', padding: '0.5rem 1rem' }}>
                        <FaSignOutAlt style={{ marginRight: '0.5rem' }} /> Logout
                    </button>
                </div>
            </nav>

            <main className="container" style={{ padding: '2rem 1rem' }}>
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.5 }}
                >
                    {error && <div className="error-msg">{error}</div>}

                    {!profile && isEditing && (
                        <div className="card" style={{ maxWidth: '800px', margin: '0 auto' }}>
                            <h2 className="title" style={{ fontSize: '1.5rem', marginBottom: '1.5rem' }}>Create Your Patient Profile</h2>
                            <ProfileForm
                                formData={formData}
                                setFormData={setFormData}
                                handleSave={handleSave}
                                isNew={true}
                            />
                        </div>
                    )}

                    {profile && !isEditing && (
                        <div style={{ maxWidth: '800px', margin: '0 auto' }}>
                            <div className="card" style={{ marginBottom: '2rem' }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1.5rem' }}>
                                    <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
                                        <FaUserCircle size={60} color="var(--primary-light)" style={{ color: 'var(--primary)' }} />
                                        <div>
                                            <h2 style={{ fontSize: '1.5rem', fontWeight: 600 }}>{profile.full_name}</h2>
                                            <p style={{ color: 'var(--text-muted)' }}>Patient ID: {profile.id.substring(0, 8)}...</p>
                                        </div>
                                    </div>
                                    <button onClick={() => setIsEditing(true)} className="btn btn-outline">
                                        <FaEdit style={{ marginRight: '0.5rem' }} /> Edit Profile
                                    </button>
                                </div>

                                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1.5rem' }}>
                                    <InfoItem label="Date of Birth" value={new Date(profile.date_of_birth).toLocaleDateString()} />
                                    <InfoItem label="Gender" value={profile.gender} />
                                    <InfoItem label="Phone" value={profile.phone} />
                                    <InfoItem label="Address" value={profile.address} />
                                </div>
                            </div>

                            {/* Share Section */}
                            <div className="card" style={{ background: '#ECFDF5', borderColor: '#D1FAE5' }}>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '1rem' }}>
                                    <div style={{ background: 'var(--white)', padding: '0.75rem', borderRadius: '50%' }}>
                                        <FaShareAlt size={24} color="var(--primary)" />
                                    </div>
                                    <div>
                                        <h3 style={{ fontSize: '1.25rem', fontWeight: 600, color: 'var(--primary-dark)' }}>Share Medical History</h3>
                                        <p style={{ fontSize: '0.875rem', color: 'var(--primary-dark)' }}>Grant temporary access to doctors via link or QR code.</p>
                                    </div>
                                </div>

                                {!shareData ? (
                                    <button onClick={generateShareLink} className="btn btn-primary">
                                        Generate Share Link
                                    </button>
                                ) : (
                                    <motion.div
                                        initial={{ opacity: 0, height: 0 }}
                                        animate={{ opacity: 1, height: 'auto' }}
                                        style={{ background: 'var(--white)', padding: '1.5rem', borderRadius: 'var(--radius)', marginTop: '1rem' }}
                                    >
                                        <label className="input-label" style={{ fontSize: '0.875rem' }}>Public Share Link</label>
                                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                                            <input
                                                readOnly
                                                value={`${window.location.origin}/share/${shareData.share_token}`}
                                                className="input-field"
                                                style={{ background: '#F9FAFB' }}
                                            />
                                            <button
                                                className="btn btn-primary"
                                                onClick={() => navigator.clipboard.writeText(`${window.location.origin}/share/${shareData.share_token}`)}
                                            >
                                                Copy
                                            </button>
                                        </div>
                                        <div style={{ marginTop: '1rem', display: 'flex', justifyContent: 'center' }}>
                                            {/* Placeholder for QR Code - In a real app use a QR lib */}
                                            <div style={{ textAlign: 'center', color: 'var(--text-muted)', fontSize: '0.875rem' }}>
                                                <FaQrcode size={120} color="var(--text-main)" />
                                                <p style={{ marginTop: '0.5rem' }}>Scan to view</p>
                                            </div>
                                        </div>
                                    </motion.div>
                                )}
                            </div>
                        </div>
                    )}

                    {profile && isEditing && (
                        <div className="card" style={{ maxWidth: '800px', margin: '0 auto' }}>
                            <h2 className="title" style={{ fontSize: '1.5rem', marginBottom: '1.5rem' }}>Edit Profile</h2>
                            <ProfileForm
                                formData={formData}
                                setFormData={setFormData}
                                handleSave={handleSave}
                                onCancel={() => setIsEditing(false)}
                            />
                        </div>
                    )}
                </motion.div>
            </main>
        </div>
    );
};

const InfoItem = ({ label, value }) => (
    <div>
        <span style={{ display: 'block', fontSize: '0.875rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>{label}</span>
        <span style={{ fontWeight: 500 }}>{value || 'Not set'}</span>
    </div>
);

const ProfileForm = ({ formData, setFormData, handleSave, isNew, onCancel }) => {
    const handleChange = (e) => {
        setFormData({ ...formData, [e.target.name]: e.target.value });
    };

    return (
        <form onSubmit={handleSave}>
            <div className="input-group">
                <label className="input-label">Full Name</label>
                <input name="full_name" value={formData.full_name} onChange={handleChange} className="input-field" required />
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div className="input-group">
                    <label className="input-label">Date of Birth</label>
                    <input type="date" name="date_of_birth" value={formData.date_of_birth} onChange={handleChange} className="input-field" required />
                </div>
                <div className="input-group">
                    <label className="input-label">Gender</label>
                    <select name="gender" value={formData.gender} onChange={handleChange} className="input-field">
                        <option value="Male">Male</option>
                        <option value="Female">Female</option>
                        <option value="Other">Other</option>
                    </select>
                </div>
            </div>
            <div className="input-group">
                <label className="input-label">Phone</label>
                <input name="phone" value={formData.phone} onChange={handleChange} className="input-field" />
            </div>
            <div className="input-group">
                <label className="input-label">Address</label>
                <textarea name="address" value={formData.address} onChange={handleChange} className="input-field" rows="3" />
            </div>
            <div style={{ display: 'flex', gap: '1rem', justifyContent: 'flex-end' }}>
                {!isNew && (
                    <button type="button" onClick={onCancel} className="btn btn-outline">Cancel</button>
                )}
                <button type="submit" className="btn btn-primary">Save Profile</button>
            </div>
        </form>
    );
};

export default Profile;
